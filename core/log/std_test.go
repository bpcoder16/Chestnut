package log

import (
	"bytes"
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestStdLogger(t *testing.T) {
	logger := DefaultLogger
	logger = With(logger, "caller", DefaultCaller, "ts", DefaultTimestamp)

	_ = logger.Log(LevelDebug, "msg", "test debug")
	_ = logger.Log(LevelInfo, "msg", "test info")
	_ = logger.Log(LevelWarn, "msg", "test Warn")
	_ = logger.Log(LevelError, "msg", "test Error")
	_ = logger.Log(LevelFatal, "msg", "test Fatal")

	logger2 := DefaultLogger
	_ = logger2.Log(LevelDebug)
	_ = logger2.Log(LevelDebug, "msg2", "test debug")
}

func TestStdLoggerLog(t *testing.T) {
	var b bytes.Buffer
	logger := NewStdLogger(&b)

	var eg errgroup.Group
	eg.Go(func() error {
		return logger.Log(LevelInfo, "msg", "a", "k", "v")
	})
	eg.Go(func() error {
		return logger.Log(LevelInfo, "msg", "a", "k", "v")
	})
	err := eg.Wait()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s := b.String(); s != "INFO msg=a k=v\nINFO msg=a k=v\n" {
		t.Fatalf("unexpected output: %s", s)
	}
}
