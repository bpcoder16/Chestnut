package esclientv7

import (
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Logger struct {
	*log.Helper
	EnableDebug bool
}

func (l *Logger) LogRoundTrip(req *http.Request, resp *http.Response, err error, _ time.Time, duration time.Duration) error {
	// 打印请求 Body
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
	}
	var respBody []byte
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
	}

	l.Helper.WithContext(req.Context()).DebugW(
		"costTime", strconv.FormatInt(duration.Milliseconds(), 10)+"ms",
		"URL", req.URL.String(),
		"method", req.Method,
		"requestBody", string(reqBody),
		"responseBody", string(respBody),
		"error", err,
	)
	return nil
}

func (l *Logger) RequestBodyEnabled() bool {
	return l.EnableDebug
}

func (l *Logger) ResponseBodyEnabled() bool {
	return l.EnableDebug
}
