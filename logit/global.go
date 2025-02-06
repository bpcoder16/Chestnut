package logit

import (
	"context"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"sync"
)

var global = &loggerAppliance{}

type loggerAppliance struct {
	lock   sync.Mutex
	helper *log.Helper
}

func init() {
	global.SetLogger(log.DefaultLogger)
}

func GetGlobalHelper() *log.Helper {
	return global.helper
}

func (a *loggerAppliance) SetLogger(in log.Logger) {
	a.lock.Lock()
	defer a.lock.Unlock()
	switch v := in.(type) {
	case *log.Helper:
		a.helper = v
	default:
		a.helper = log.NewHelper(v)
	}
}

func SetLogger(logger log.Logger) {
	global.SetLogger(logger)
}

func Log(level log.Level, keyValues ...interface{}) error {
	return global.helper.Log(level, keyValues...)
}

func Context(ctx context.Context) *log.Helper {
	return global.helper.WithContext(ctx)
}

func Debug(a ...interface{}) {
	global.helper.Debug(a...)
}

func DebugF(format string, a ...interface{}) {
	global.helper.DebugF(format, a...)
}

func DebugW(keyValues ...interface{}) {
	global.helper.DebugW(keyValues...)
}

func Info(a ...interface{}) {
	global.helper.Info(a...)
}

func InfoF(format string, a ...interface{}) {
	global.helper.InfoF(format, a...)
}

func InfoW(keyValues ...interface{}) {
	global.helper.InfoW(keyValues...)
}

func Warn(a ...interface{}) {
	global.helper.Warn(a...)
}

func WarnF(format string, a ...interface{}) {
	global.helper.WarnF(format, a...)
}

func WarnW(keyValues ...interface{}) {
	global.helper.WarnW(keyValues...)
}

func Error(a ...interface{}) {
	global.helper.Error(a...)
}

func ErrorF(format string, a ...interface{}) {
	global.helper.ErrorF(format, a...)
}

func ErrorW(keyValues ...interface{}) {
	global.helper.ErrorW(keyValues...)
}

func Fatal(a ...interface{}) {
	global.helper.Fatal(a...)
}

func FatalF(format string, a ...interface{}) {
	global.helper.FatalF(format, a...)
}

func FatalW(keyValues ...interface{}) {
	global.helper.FatalW(keyValues...)
}
