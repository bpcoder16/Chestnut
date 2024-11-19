package log

import (
	"context"
	"testing"
)

func TestLogInfo(t *testing.T) {
	logger := WithContext(context.Background(), DefaultLogger)
	logger = With(DefaultLogger, "ts", DefaultTimestamp)
	logger = With(logger, "caller", DefaultCaller)
	_ = logger.Log(LevelInfo, "key", "value")
}
