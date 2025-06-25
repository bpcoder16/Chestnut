package gin

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/gin-gonic/gin"
	"io"
	gLog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func defaultLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		begin := time.Now()

		ctx.Set(log.DefaultMessageKey, "HTTP")
		ctx.Set(log.DefaultLogIdKey, utils.UniqueID())

		reqBody := generateRequestBody(ctx)

		writer := &responseWriter{
			ResponseWriter: ctx.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		ctx.Writer = writer

		ctx.Next()

		elapsed := time.Since(begin)

		logit.Context(ctx).InfoW(
			"userId", ctx.GetInt(log.DefaultUserIdKey),
			"costTime", utils.ShowDurationString(elapsed),
			"clientIP", ctx.ClientIP(),
			"method", ctx.Request.Method,
			"uri", ctx.Request.URL.Path,
			"rawQuery", ctx.Request.URL.RawQuery,
			"header", filterHeader(ctx.Request.Header),
			"reqBody", filterBody(reqBody),
			"statusCode", ctx.Writer.Status(),
			"response", filterBody(writer.body.Bytes()),
		)
	}
}

func generateRequestBody(ctx *gin.Context) []byte {
	body, _ := ctx.GetRawData()                            // 读取 request body 的内容
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // 创建 io.ReadCloser 对象传给 request body
	return body                                            // 返回 request body 的值
}

// 自定义一个结构体，实现 gin.ResponseWriter interface
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func recoveryWithWriter(out io.Writer) gin.HandlerFunc {
	var logger *gLog.Logger
	if out != nil {
		logger = gLog.New(out, "\n\n\x1b[31m", gLog.LstdFlags)
	}
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				if logger != nil {
					stackBytes := stack(3)
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if current[0] == "Authorization" {
							headers[idx] = current[0] + ": *"
						}
					}
					if brokenPipe {
						logger.Printf("%s\n%s%s", err, string(httpRequest), reset)
					} else {
						logger.Printf("[Recovery] %s panic recovered:\n%s\n%s%s",
							time.Now().Format(time.DateTime), err, stackBytes, reset)
					}
				}

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func corsPreCheckRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.Header("Access-Control-Allow-Origin", "*")
			// 允许的 HTTP 方法
			ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			// 允许的请求头
			ctx.Header("Access-Control-Allow-Headers", "*")
			// 预检请求的缓存时间
			ctx.Header("Access-Control-Max-Age", "86400")

			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		// 允许的 HTTP 方法
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// 允许的请求头
		ctx.Header("Access-Control-Allow-Headers", "*")
		// 预检请求的缓存时间
		ctx.Header("Access-Control-Max-Age", "86400")

		ctx.Next()
	}
}

func SignAuthMiddleware(secretKeyMap map[string]string, timeWindow time.Duration, toStringFunc func(any) string) gin.HandlerFunc {
	buildSortedParamString := func(params map[string]any, toStringFunc func(any) string) string {
		var keys []string
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		pairs := url.Values{}
		for _, key := range keys {
			pairs.Set(key, toStringFunc(params[key]))
		}

		return pairs.Encode()
	}
	absDuration := func(d time.Duration) time.Duration {
		if d < 0 {
			return -d
		}
		return d
	}

	return func(ctx *gin.Context) {
		appID := ctx.GetHeader("X-APP-ID")
		signature := ctx.GetHeader("X-Signature")
		timestampStr := ctx.GetHeader("X-Timestamp")

		var secretKey string
		var isExist bool
		if secretKey, isExist = secretKeyMap[appID]; !isExist {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "验签不通过",
			})
			ctx.Abort()
			return
		}

		if signature == "" || timestampStr == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "缺少签名或时间戳",
			})
			ctx.Abort()
			return
		}

		// 时间戳校验
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "时间戳无效或超时",
			})
			ctx.Abort()
			return
		}
		reqTime := time.Unix(timestamp, 0)
		if absDuration(time.Now().Sub(reqTime)) > timeWindow {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "时间戳无效或超时",
			})
			ctx.Abort()
			return
		}

		// 读取原始 body
		reqBody := generateRequestBody(ctx)

		// 解析 JSON 并排序为字符串
		var params map[string]interface{}
		if errJ := json.Unmarshal(reqBody, &params); errJ != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":  http.StatusBadRequest,
				"error": "JSON 解析失败",
			})
			ctx.Abort()
			return
		}
		// 加入时间戳
		params["timestamp"] = timestamp
		signStr := buildSortedParamString(params, toStringFunc)

		// 计算 HMAC-SHA256 签名
		mac := hmac.New(sha256.New, []byte(secretKey))
		mac.Write([]byte(signStr))
		expectedSign := hex.EncodeToString(mac.Sum(nil))

		if expectedSign != signature {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "验签不通过",
			})
			logit.Context(ctx).ErrorW(
				"signStr", signStr,
				"expectedSign", expectedSign,
			)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
