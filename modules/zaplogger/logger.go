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
//		log.FilterLevel(log.LevelDebug),
//		log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
//			return true
//		}),
//	)

func GetZapLogger(debugWriter, infoWriter, warnErrorFatalWriter io.Writer, caller log.Valuer, opts ...log.FilterOption) log.Logger {
	kv := make([]interface{}, 0, 8)
	kv = append(kv,
		log.DefaultMessageKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				msg := ctx.Value(log.DefaultMessageKey)
				if msg == nil {
					return "None"
				}
				return msg
			}
		}(),
		log.DefaultLogIdKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				logId := ctx.Value(log.DefaultLogIdKey)
				if logId == nil {
					return "None"
				}
				return logId
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
		log.DefaultWebSocketUUIDKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				msg := ctx.Value(log.DefaultWebSocketUUIDKey)
				if msg == nil {
					return "None"
				}
				return msg
			}
		}(),
		log.DefaultWebSocketLogIdKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				msg := ctx.Value(log.DefaultWebSocketLogIdKey)
				if msg == nil {
					return "None"
				}
				return msg
			}
		}(),
		log.DefaultCronActionKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				msg := ctx.Value(log.DefaultCronActionKey)
				if msg == nil {
					return "None"
				}
				return msg
			}
		}(),
		log.DefaultWebSocketPathKey,
		func() log.Valuer {
			return func(ctx context.Context) interface{} {
				msg := ctx.Value(log.DefaultWebSocketPathKey)
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
