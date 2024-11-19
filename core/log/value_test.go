package log

import (
	"context"
	"testing"
)

func TestValue(t *testing.T) {
	logger := With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller)
	_ = logger.Log(LevelInfo, "msg", "Hello World")

	logger = DefaultLogger
	logger = With(logger)
	_ = logger.Log(LevelDebug, "msg", "Hello World")

	var v1 interface{}
	got := Value(context.Background(), v1)
	if got != v1 {
		t.Errorf("got %#v, expected %#v", got, v1)
	}

	var v2 Valuer = func(ctx context.Context) interface{} {
		return 3
	}

	got = Value(context.Background(), v2)
	res := got.(int)
	if res != 3 {
		t.Errorf("got %#v, expected %#v", res, 3)
	}
}
