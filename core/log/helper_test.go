package log

import "testing"

func TestHelper(t *testing.T) {
	logger := NewHelper(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
	)

	_ = logger.Log(LevelDebug, "msg", "test debug")
	logger.Debug("test debug")
	logger.DebugF("test %s", "debug")
	logger.DebugW("logit", "test debug")

	logger.Warn("test warn")
	logger.WarnF("test %s", "warn")
	logger.WarnW("logit", "test debug")
}

func TestHelperWithMsgKey(t *testing.T) {
	logger := NewHelper(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
		WithMessageKey("message"),
	)

	logger.DebugF("test %s", "debug")
	logger.DebugW("logit", "test debug")
}

func TestHelperLevel(t *testing.T) {
	logger := NewHelper(DefaultLogger)

	logger.Debug("Debug")
	logger.Info("Info")
	logger.InfoF("test %s", "info")
	logger.Warn("Warn")
	logger.Error("Error")
	logger.ErrorF("test %s", "error")
	logger.ErrorW("logit", "test error")
}
