package gin

import (
	"bytes"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/gin-gonic/gin"
	"io"
	gLog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
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
			ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, If-Match, If-Modified-Since, If-None-Match, If-Unmodified-Since, X-CSRF-TOKEN, X-Requested-With, Token, equipmentType, pseudoUniqueID")
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
		ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, If-Match, If-Modified-Since, If-None-Match, If-Unmodified-Since, X-CSRF-TOKEN, X-Requested-With, Token, equipmentType, pseudoUniqueID")
		// 预检请求的缓存时间
		ctx.Header("Access-Control-Max-Age", "86400")

		ctx.Next()
	}
}
