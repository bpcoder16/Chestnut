package zaplogger

import (
	"context"
	"github.com/bpcoder16/Chestnut/contrib/log/zap"
	"github.com/bpcoder16/Chestnut/core/log"
	"io"
)

// 	zaplogger.GetZapLogger(
//		"/Users/bpcoder/git/one",
//		"one",
//		log.FilterKey("password"),
//		log.FilterValue("bpcoder"),
//		log.FilterLevel(log.LevelDebug),
//		log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
//			return true
//		}),
//	)

func GetZapLogger(debugWriter, infoWriter, warnErrorFatalWriter io.Writer, caller log.Valuer, opts ...log.FilterOption) log.Logger {
	kv := make([]interface{}, 0, 8)
	kv = append(kv, log.DefaultLogIdKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				logId := ctx.Value(log.DefaultLogIdKey)
				if logId == nil {
					return "NotSetLogId"
				}
				return logId
			}
		}(),
		log.DefaultMessageKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				msg := ctx.Value(log.DefaultMessageKey)
				if msg == nil {
					return "Default"
				}
				return msg
			}
		}(),
		log.DefaultDownstreamKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				msg := ctx.Value(log.DefaultDownstreamKey)
				if msg == nil {
					return "None"
				}
				return msg
			}
		}(),
		log.DefaultConcurrencyLogIdKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				msg := ctx.Value(log.DefaultConcurrencyLogIdKey)
				if msg == nil {
					return "None"
				}
				return msg
			}
		}(),
	)
	if caller != nil {
		kv = append(kv, log.DefaultCallerKey, caller)
	}

	return log.NewFilter(
		log.With(
			zap.NewLogger(debugWriter, infoWriter, warnErrorFatalWriter),
			kv...,
		),
		opts...,
	)
}
