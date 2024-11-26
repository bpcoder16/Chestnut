package gin

import (
	"bytes"
	"fmt"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		ctx.Set(log.DefaultLogIdKey, uuid.New().String())

		reqBody := generateRequestBody(ctx)

		writer := &responseWriter{
			ResponseWriter: ctx.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		ctx.Writer = writer

		ctx.Next()

		elapsed := time.Since(begin)

		logit.Context(ctx).InfoW(
			"costTime", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
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
