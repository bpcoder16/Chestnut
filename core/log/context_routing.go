package log

import (
	"context"
	"fmt"
)

type ContextRoutingLogger struct {
	defaultLogger Logger
	cronLogger    Logger
}

func NewContextRoutingLogger(defaultLogger, cronLogger Logger) *ContextRoutingLogger {
	return &ContextRoutingLogger{
		defaultLogger: defaultLogger,
		cronLogger:    cronLogger,
	}
}

func (l *ContextRoutingLogger) Log(level Level, keyValues ...interface{}) error {
	if IsCronKeyValues(keyValues...) {
		return l.cronLogger.Log(level, keyValues...)
	}
	return l.defaultLogger.Log(level, keyValues...)
}

func IsCronContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if ctx.Value(DefaultCronActionKey) != nil {
		return true
	}
	msg, _ := ctx.Value(DefaultMessageKey).(string)
	return msg == DefaultCronMessageValue
}

func IsCronKeyValues(keyValues ...interface{}) bool {
	for i := 0; i < len(keyValues); i += 2 {
		v := i + 1
		if v >= len(keyValues) {
			break
		}
		switch fmt.Sprint(keyValues[i]) {
		case DefaultCronActionKey:
			if keyValues[v] != nil && fmt.Sprint(keyValues[v]) != "None" {
				return true
			}
		case DefaultMessageKey:
			if fmt.Sprint(keyValues[v]) == DefaultCronMessageValue {
				return true
			}
		}
	}
	return false
}
