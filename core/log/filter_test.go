package log

import (
	"bytes"
	"context"
	"golang.org/x/sync/errgroup"
	"strings"
	"testing"
	"time"
)

func testFilterFunc(_ Level, keyValues ...interface{}) bool {
	for i := 0; i < len(keyValues); i += 2 {
		if keyValues[i] == "password" {
			keyValues[i+1] = fuzzyStr
		}
	}
	return false
}

func TestFilterAll(t *testing.T) {
	logger := NewHelper(NewFilter(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
		FilterLevel(LevelDebug),
		FilterKey("username"),
		FilterValue("hello"),
		FilterFunc(testFilterFunc),
	))

	_ = logger.Log(LevelDebug, "msg", "test debug")
	logger.Info("hello")
	logger.InfoF("hello %s", "world")
	logger.InfoW("username", "water")
	logger.InfoW("password", "123456")
	logger.Warn("warn logit")
}

func TestFilterLevel(t *testing.T) {
	logger := NewHelper(NewFilter(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
		FilterLevel(LevelWarn),
	))
	_ = logger.Log(LevelDebug, "msg", "test debug")
	logger.Debug("Debug")
	logger.Info("Info")
	logger.Warn("Warn")
	logger.Error("Error")
	//zaplogger.Fatal("Fatal")
}

func TestFilerCaller(t *testing.T) {
	logger := NewFilter(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
	)

	_ = logger.Log(LevelDebug, "msg", "test debug")

	helper := NewHelper(logger)
	_ = helper.Log(LevelDebug, "msg1", "test debug")
}

func TestFilterKey(t *testing.T) {
	logger := NewHelper(NewFilter(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
		FilterKey("password"),
	))
	logger.DebugW("password", "123456")
}

func TestFilterValue(t *testing.T) {
	logger := NewHelper(NewFilter(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
		FilterValue("debug"),
	))
	logger.DebugF("test %s", "debug")
	logger.Debug("debug", "debug")
	logger.Debug("debug")
	logger.DebugW("key", "debug")
}

func TestFilterFunc(t *testing.T) {
	logger := NewHelper(NewFilter(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
		FilterFunc(testFilterFunc),
	))

	logger.Debug("debug level")
	logger.InfoW("password", "123456")
}

func BenchmarkFilterAll(b *testing.B) {
	log := NewHelper(NewFilter(
		With(DefaultLogger, "ts", DefaultTimestamp, "caller", DefaultCaller),
		FilterKey("username"),
		FilterValue("Water"),
		FilterFunc(testFilterFunc),
	))
	for i := 0; i < b.N; i++ {
		log.InfoW("username", "1111", "password", "123456", "msg", "Water")
	}
}

func testFilterFuncWithLoggerPrefix(level Level, keyValues ...interface{}) bool {
	if level == LevelWarn {
		return true
	}
	for i := 0; i < len(keyValues); i += 2 {
		if keyValues[i] == "prefix" {
			return true
		}
		if keyValues[i] == "filtered" {
			keyValues[i+1] = fuzzyStr
		}
	}
	return false
}

func TestFilterFuncLoggerPrefix(t *testing.T) {
	buf := new(bytes.Buffer)
	tests := []struct {
		logger Logger
		want   string
	}{
		{
			logger: NewFilter(
				With(NewStdLogger(buf), "caller", "caller", "prefix", "whatever"),
				FilterFunc(testFilterFuncWithLoggerPrefix),
			),
			want: "",
		},
		{
			logger: NewFilter(
				With(NewStdLogger(buf), "caller", "caller"),
				FilterFunc(testFilterFuncWithLoggerPrefix),
			),
			want: "INFO caller=caller msg=msg filtered=***\n",
		},
		{
			logger: NewFilter(
				With(NewStdLogger(buf)),
				FilterFunc(testFilterFuncWithLoggerPrefix),
			),
			want: "INFO msg=msg filtered=***\n",
		},
	}

	for _, tt := range tests {
		err := tt.logger.Log(LevelInfo, "msg", "msg", "filtered", "true")
		if err != nil {
			t.Fatal("err should be nil")
		}
		got := buf.String()
		if got != tt.want {
			t.Fatalf("filter zaplogger want %s, got %s", tt.want, got)
		}
		buf.Reset()
	}
}

func TestFilterWithContext(t *testing.T) {
	ctxKey := struct{}{}
	ctxValue := "filter test value"

	v1 := func() Valuer {
		return func(ctx context.Context) interface{} {
			return ctx.Value(ctxKey)
		}
	}

	info := new(bytes.Buffer)

	logger := NewFilter(
		With(NewStdLogger(info), "request_id", v1()),
		FilterLevel(LevelError),
	)

	ctx := context.WithValue(context.Background(), ctxKey, ctxValue)

	_ = WithContext(ctx, logger).Log(LevelInfo, "kind", "test")
	if info.String() != "" {
		t.Error("filter is not working")
	}

	_ = WithContext(ctx, logger).Log(LevelError, "kind", "test")
	if !strings.Contains(info.String(), ctxValue) {
		t.Error("don't read ctx value")
	}
}

type traceIDKey struct{}

func setTraceID(ctx context.Context, tid string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, tid)
}

func traceIDValuer() Valuer {
	return func(ctx context.Context) any {
		if ctx == nil {
			return ""
		}
		if tid := ctx.Value(traceIDKey{}); tid != nil {
			return tid
		}
		return ""
	}
}

func TestFilterWithContextConcurrent(t *testing.T) {
	buf := new(bytes.Buffer)
	pCtx := context.Background()
	l := NewFilter(
		With(NewStdLogger(buf), "trace-id", traceIDValuer()),
		FilterLevel(LevelInfo),
	)

	var eg errgroup.Group
	eg.Go(func() error {
		time.Sleep(time.Second)
		NewHelper(l).Info("done1")
		return nil
	})

	eg.Go(func() error {
		tid := "world"
		ctx := setTraceID(pCtx, tid)
		NewHelper(WithContext(ctx, l)).Info("done2")
		return nil
	})

	err := eg.Wait()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "INFO trace-id=world msg=done2\nINFO trace-id= msg=done1\n"
	if got := buf.String(); got != expected {
		t.Errorf("got: %s, want: %s", got, expected)
	}
}
