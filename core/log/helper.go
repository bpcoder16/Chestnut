package log

import (
	"context"
	"fmt"
	"os"
)

const (
	DefaultMessageKey    = "type"
	DefaultDownstreamKey = "downstream"
	DefaultLogIdKey      = "logId"
	DefaultCallerKey     = "caller"
)

type HelperOption func(*Helper)

type Helper struct {
	logger  Logger
	msgKey  string
	sprint  func(...interface{}) string
	sprintf func(format string, a ...interface{}) string
}

// WithMessageKey with message key.
func WithMessageKey(k string) HelperOption {
	return func(opts *Helper) {
		opts.msgKey = k
	}
}

// WithSprint with sprint
func WithSprint(sprint func(...interface{}) string) HelperOption {
	return func(opts *Helper) {
		opts.sprint = sprint
	}
}

// WithSprintf with sprintf
func WithSprintf(sprintf func(format string, a ...interface{}) string) HelperOption {
	return func(opts *Helper) {
		opts.sprintf = sprintf
	}
}

func NewHelper(logger Logger, opts ...HelperOption) *Helper {
	options := &Helper{
		msgKey:  DefaultMessageKey, // default message key
		logger:  logger,
		sprint:  fmt.Sprint,
		sprintf: fmt.Sprintf,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

func (h *Helper) WithContext(ctx context.Context) *Helper {
	return &Helper{
		msgKey:  h.msgKey,
		logger:  WithContext(ctx, h.logger),
		sprint:  h.sprint,
		sprintf: h.sprintf,
	}
}

func (h *Helper) WithValues(kv ...interface{}) *Helper {
	return &Helper{
		msgKey:  h.msgKey,
		logger:  With(h.logger, kv...),
		sprint:  h.sprint,
		sprintf: h.sprintf,
	}
}

func (h *Helper) Log(level Level, keyValues ...interface{}) error {
	return h.logger.Log(level, keyValues...)
}

func (h *Helper) Debug(a ...interface{}) {
	_ = h.logger.Log(LevelDebug, h.msgKey, h.sprint(a...))
}

func (h *Helper) DebugF(format string, a ...interface{}) {
	_ = h.logger.Log(LevelDebug, h.msgKey, h.sprintf(format, a...))
}

func (h *Helper) DebugW(keyValues ...interface{}) {
	_ = h.logger.Log(LevelDebug, keyValues...)
}

func (h *Helper) Info(a ...interface{}) {
	_ = h.logger.Log(LevelInfo, h.msgKey, h.sprint(a...))
}

func (h *Helper) InfoF(format string, a ...interface{}) {
	_ = h.logger.Log(LevelInfo, h.msgKey, h.sprintf(format, a...))
}

func (h *Helper) InfoW(keyValues ...interface{}) {
	_ = h.logger.Log(LevelInfo, keyValues...)
}

func (h *Helper) Warn(a ...interface{}) {
	_ = h.logger.Log(LevelWarn, h.msgKey, h.sprint(a...))
}

func (h *Helper) WarnF(format string, a ...interface{}) {
	_ = h.logger.Log(LevelWarn, h.msgKey, h.sprintf(format, a...))
}

func (h *Helper) WarnW(keyValues ...interface{}) {
	_ = h.logger.Log(LevelWarn, keyValues...)
}

func (h *Helper) Error(a ...interface{}) {
	_ = h.logger.Log(LevelError, h.msgKey, h.sprint(a...))
}

func (h *Helper) ErrorF(format string, a ...interface{}) {
	_ = h.logger.Log(LevelError, h.msgKey, h.sprintf(format, a...))
}

func (h *Helper) ErrorW(keyValues ...interface{}) {
	_ = h.logger.Log(LevelError, keyValues...)
}

func (h *Helper) Fatal(a ...interface{}) {
	_ = h.logger.Log(LevelFatal, h.msgKey, h.sprint(a...))
	os.Exit(1)
}

func (h *Helper) FatalF(format string, a ...interface{}) {
	_ = h.logger.Log(LevelFatal, h.msgKey, h.sprintf(format, a...))
	os.Exit(1)
}

func (h *Helper) FatalW(keyValues ...interface{}) {
	_ = h.logger.Log(LevelFatal, keyValues...)
	os.Exit(1)
}
