package bootstrap

import (
	"context"
	"github.com/bpcoder16/Chestnut/appconfig"
	env2 "github.com/bpcoder16/Chestnut/appconfig/env"
	"github.com/bpcoder16/Chestnut/clickhouse"
	"github.com/bpcoder16/Chestnut/contrib/aliyun/oss"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/bpcoder16/Chestnut/lru"
	"github.com/bpcoder16/Chestnut/modules/zaplogger"
	"github.com/bpcoder16/Chestnut/mongodb"
	"github.com/bpcoder16/Chestnut/mysql"
	"github.com/bpcoder16/Chestnut/redis"
	"github.com/bpcoder16/Chestnut/resty"
	"io"
	"path"
)

func MustInit(ctx context.Context, config *appconfig.AppConfig, funcList ...func(ctx context.Context, debugWriter, infoWriter, warnErrorFatalWriter io.Writer)) {
	var debugWriter, infoWriter, warnErrorFatalWriter io.Writer
	if config.NotUseRotateLog {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetStandardWriters(config.LogDir, env2.AppName(), env2.AppName())
	} else {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetFileRotateLogWriters(config.LogDir, env2.AppName(), env2.AppName())
	}
	if config.StdRedirectFileSupport {
		zaplogger.StdRedirectFile(config.LogDir)
	}
	initLoggers(ctx, config, debugWriter, infoWriter, warnErrorFatalWriter)

	if config.AliyunOSSSupport {
		initAliyunOSS()
	}
	if config.DefaultMongoDBSupport {
		initMongoDB(ctx, debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.DefaultRedisSupport {
		initRedis(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.UseLRUCache {
		initUseLRUCache(ctx, debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.DefaultMySQLSupport {
		initMySQL(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.DefaultClickhouseSupport {
		initClickhouse(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	initHTTPClient()
	for _, fn := range funcList {
		fn(ctx, debugWriter, infoWriter, warnErrorFatalWriter)
	}
}

func initLoggers(_ context.Context, config *appconfig.AppConfig, debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	logit.SetLogger(zaplogger.GetZapLogger(
		debugWriter, infoWriter, warnErrorFatalWriter,
		log.FileWithLineNumCaller(),
		log.FilterKey(config.FilterKeys...),
		log.FilterValue(config.FilterValues...),
		log.FilterLevel(func() log.Level {
			if env2.RunMode() == env2.RunModeRelease {
				return log.LevelInfo
			}
			return log.LevelDebug
		}()),
		//log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
		//	return false
		//}),
	))
}

func initAliyunOSS() {
	oss.InitAliyunOSS(path.Join(env2.RootPath(), "conf/aliyun.json"))
}

func initMongoDB(ctx context.Context, debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	mongodb.SetManager(ctx, path.Join(env2.RootPath(), "conf/mongodb.json"), log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			log.FileWithLineNumCaller(),
			log.FilterLevel(func() log.Level {
				if env2.RunMode() == env2.RunModeRelease {
					return log.LevelInfo
				}
				return log.LevelDebug
			}()),
			//log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
			//	return false
			//}),
		),
	))
}

func initUseLRUCache(_ context.Context, debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	lru.SetManager(path.Join(env2.RootPath(), "conf/lru.json"), log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			log.FileWithLineNumCaller(),
			log.FilterLevel(func() log.Level {
				if env2.RunMode() == env2.RunModeRelease {
					return log.LevelInfo
				}
				return log.LevelDebug
			}()),
			//log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
			//	return false
			//}),
		),
	))
}

func initRedis(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	redis.SetManager(path.Join(env2.RootPath(), "conf/redis.json"), log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			log.FileWithLineNumCallerRedis(),
			log.FilterLevel(func() log.Level {
				if env2.RunMode() == env2.RunModeRelease {
					return log.LevelInfo
				}
				return log.LevelDebug
			}()),
			//log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
			//	return false
			//}),
		),
	))
}

func initMySQL(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	mysql.SetManager(env2.RootPath()+"/conf/mysql.json", log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			nil,
			log.FilterLevel(func() log.Level {
				if env2.RunMode() == env2.RunModeRelease {
					return log.LevelInfo
				}
				return log.LevelDebug
			}()),
			//log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
			//	return false
			//}),
		),
	))
}

func initClickhouse(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	clickhouse.SetManager(env2.RootPath()+"/conf/clickhouse.json", log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			nil,
			log.FilterLevel(func() log.Level {
				if env2.RunMode() == env2.RunModeRelease {
					return log.LevelInfo
				}
				return log.LevelDebug
			}()),
			//log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
			//	return false
			//}),
		),
	))
}

func initHTTPClient() {
	resty.SetClient()
}
