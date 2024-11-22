package bootstrap

import (
	"context"
	"github.com/bpcoder16/Chestnut/clickhouse"
	"github.com/bpcoder16/Chestnut/contrib/aliyun/oss"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/bpcoder16/Chestnut/modules/appconfig"
	"github.com/bpcoder16/Chestnut/modules/appconfig/env"
	"github.com/bpcoder16/Chestnut/modules/zaplogger"
	"github.com/bpcoder16/Chestnut/mongodb"
	"github.com/bpcoder16/Chestnut/mysql"
	"github.com/bpcoder16/Chestnut/redis"
	"github.com/bpcoder16/Chestnut/resty"
	"io"
)

func MustInit(ctx context.Context, config *appconfig.AppConfig, funcList ...func(ctx context.Context, debugWriter, infoWriter, warnErrorFatalWriter io.Writer)) {
	var debugWriter, infoWriter, warnErrorFatalWriter io.Writer
	if config.NotUseRotateLog {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetStandardWriters(config.LogDir, env.AppName(), env.AppName())
	} else {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetFileRotateLogWriters(config.LogDir, env.AppName(), env.AppName())
	}
	if config.StdRedirectFileSupport {
		zaplogger.StdRedirectFile(config.LogDir)
	}
	initLoggers(ctx, config, debugWriter, infoWriter, warnErrorFatalWriter)
	if config.DefaultMySQLSupport {
		initMySQL(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.DefaultClickhouseSupport {
		initClickhouse(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.DefaultRedisSupport {
		initRedis(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.DefaultMongoDBSupport {
		initMongoDB(ctx, debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.AliyunOSSSupport {
		initAliyunOSS()
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
			if env.RunMode() == env.RunModeRelease {
				return log.LevelInfo
			}
			return log.LevelDebug
		}()),
		//log.FilterFunc(func(level log.Level, keyValues ...interface{}) bool {
		//	return false
		//}),
	))
}

func initMySQL(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	mysql.SetManager(env.RootPath()+"/conf/mysql.json", log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			nil,
			log.FilterLevel(func() log.Level {
				if env.RunMode() == env.RunModeRelease {
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
	clickhouse.SetManager(env.RootPath()+"/conf/clickhouse.json", log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			nil,
			log.FilterLevel(func() log.Level {
				if env.RunMode() == env.RunModeRelease {
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
	redis.SetManager(env.RootPath()+"/conf/redis.json", log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			log.FileWithLineNumCallerRedis(),
			log.FilterLevel(func() log.Level {
				if env.RunMode() == env.RunModeRelease {
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

func initMongoDB(ctx context.Context, debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	mongodb.SetManager(ctx, env.RootPath()+"/conf/mongodb.json", log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			log.FileWithLineNumCallerRedis(),
			log.FilterLevel(func() log.Level {
				if env.RunMode() == env.RunModeRelease {
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

func initAliyunOSS() {
	oss.InitAliyunOSS(env.RootPath() + "/conf/aliyun.json")
}

func initHTTPClient() {
	resty.SetClient()
}
