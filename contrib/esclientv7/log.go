package esclientv7

import (
	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"net/http"
	"time"
)

type Logger struct {
	*log.Helper
}

func (l *Logger) LogRoundTrip(req *http.Request, resp *http.Response, err error, startedAt time.Time, duration time.Duration) error {
	l.Helper.WithContext(req.Context()).DebugW(
		"URL", req.URL.String(),
		"method", req.Method,
		//"requestBody", req.Body,
	)
	return nil
}

func (l *Logger) RequestBodyEnabled() bool {
	return env.RunMode() != env.RunModeRelease
}

func (l *Logger) ResponseBodyEnabled() bool {
	return env.RunMode() != env.RunModeRelease
}
