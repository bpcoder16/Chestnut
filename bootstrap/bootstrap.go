package bootstrap

import (
	"context"
	"io"
	"path"
	"time"

	"github.com/bpcoder16/Chestnut/v4/appconfig"
	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/contrib/aliyunoss"
	"github.com/bpcoder16/Chestnut/v4/contrib/geetest"
	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/default/clickhouse"
	"github.com/bpcoder16/Chestnut/v4/default/lru"
	"github.com/bpcoder16/Chestnut/v4/default/mongodb"
	"github.com/bpcoder16/Chestnut/v4/default/mysql"
	defaultNATS "github.com/bpcoder16/Chestnut/v4/default/nats"
	"github.com/bpcoder16/Chestnut/v4/default/redis"
	"github.com/bpcoder16/Chestnut/v4/default/resty"
	"github.com/bpcoder16/Chestnut/v4/default/sqlite"
	"github.com/bpcoder16/Chestnut/v4/logit"
	"github.com/bpcoder16/Chestnut/v4/modules/zaplogger"
)

func MustInit(ctx context.Context, config *appconfig.AppConfig, funcList ...func(ctx context.Context, debugWriter, infoWriter, warnErrorFatalWriter io.Writer)) {
	time.Local = env.TimeLocation()

	if config.Log.StdRedirectFileSupport {
		zaplogger.StdRedirectFile(config.Log.LogDir)
	}
	debugWriter, infoWriter, warnErrorFatalWriter := getWriters(config)
	requestWriter := getRequestWriter(config)
	cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter := getCronWriters(config)

	initLoggers(ctx, config, debugWriter, infoWriter, warnErrorFatalWriter, requestWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)

	initDefault(ctx, config, debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)

	initHTTPClient(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)

	for _, fn := range funcList {
		fn(ctx, debugWriter, infoWriter, warnErrorFatalWriter)
	}
}

func getWriters(config *appconfig.AppConfig) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	if config.Log.UseRotateLog {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetFileRotateLogWritersWithRetentionDays(config.Log.LogDir, env.AppName(), env.AppName(), config.Log.RetentionDays)
	} else {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetStandardWriters(config.Log.LogDir, env.AppName(), env.AppName())
	}
	return
}

func getRequestWriter(config *appconfig.AppConfig) io.Writer {
	if config.Log.UseRotateLog {
		return zaplogger.GetFileRotateRequestLogWriterWithRetentionDays(config.Log.LogDir, env.AppName(), env.AppName(), config.Log.RetentionDays)
	}
	return zaplogger.GetStandardRequestWriter(config.Log.LogDir, env.AppName(), env.AppName())
}

func getCronWriters(config *appconfig.AppConfig) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	if config.Log.UseRotateLog {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetFileRotateCronLogWritersWithRetentionDays(config.Log.LogDir, env.AppName(), env.AppName(), config.Log.RetentionDays)
	} else {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetStandardCronWriters(config.Log.LogDir, env.AppName(), env.AppName())
	}
	return
}

func initLoggers(
	_ context.Context,
	config *appconfig.AppConfig,
	debugWriter, infoWriter, warnErrorFatalWriter, requestWriter io.Writer,
	cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer,
) {
	logit.SetLogger(zaplogger.GetZapLogger(
		debugWriter, infoWriter, warnErrorFatalWriter,
		log.FileWithLineNumCaller(),
		log.FilterKey(config.FilterKeys...),
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
	logit.SetRequestLogger(zaplogger.GetRequestZapLogger(
		requestWriter,
		log.FileWithLineNumCaller(),
		log.FilterKey(config.FilterKeys...),
		log.FilterLevel(func() log.Level {
			if env.RunMode() == env.RunModeRelease {
				return log.LevelInfo
			}
			return log.LevelDebug
		}()),
	))
	logit.SetCronLogger(zaplogger.GetZapLogger(
		cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter,
		log.FileWithLineNumCaller(),
		log.FilterKey(config.FilterKeys...),
		log.FilterLevel(func() log.Level {
			if env.RunMode() == env.RunModeRelease {
				return log.LevelInfo
			}
			return log.LevelDebug
		}()),
	))
}

func initDefault(ctx context.Context, config *appconfig.AppConfig, debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	if config.Default.MySQLSupport {
		initMySQL(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)
	}
	if config.Default.SQLiteSupport {
		initSQLite(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)
	}
	if config.Default.ClickhouseSupport {
		initClickhouse(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)
	}
	if config.Default.RedisSupport {
		initRedis(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)
	}
	if config.Default.MongoDBSupport {
		initMongoDB(ctx, debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)
	}
	if config.Default.LRUCacheSupport {
		initLRUCache(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)
	}
	if config.Default.AliyunOSSSupport {
		initAliyunOSS()
	}
	if config.Default.GeeTestSupport {
		initGeeTest()
	}
	if config.Default.NATSSupport {
		initNATS(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter)
	}

}

func DefaultHelper(debugWriter, infoWriter, warnErrorFatalWriter io.Writer, caller log.Valuer) *log.Helper {
	return log.NewHelper(
		zaplogger.GetZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			caller,
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
	)
}

func DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer, caller log.Valuer) *log.Helper {
	return log.NewHelper(
		zaplogger.GetCronRoutingZapLogger(
			debugWriter, infoWriter, warnErrorFatalWriter,
			cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter,
			caller,
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
	)
}

func initMySQL(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	mysql.SetManager(
		path.Join(env.ConfigDirPath(), "mysql.yaml"),
		DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter, nil),
	)
}

func initSQLite(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	sqlite.SetManager(
		path.Join(env.ConfigDirPath(), "sqlite.yaml"),
		DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter, nil),
	)
}

func initClickhouse(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	clickhouse.SetManager(
		path.Join(env.ConfigDirPath(), "clickhouse.yaml"),
		DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter, nil),
	)
}

func initRedis(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	redis.SetManager(
		path.Join(env.ConfigDirPath(), "redis.yaml"),
		DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter, log.FileWithLineNumCallerRedis()),
	)
}

func initMongoDB(ctx context.Context, debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	mongodb.SetManager(
		ctx,
		path.Join(env.ConfigDirPath(), "mongodb.yaml"),
		DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter, log.FileWithLineNumCaller()),
	)
}

func initLRUCache(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	lru.SetManager(
		path.Join(env.ConfigDirPath(), "lru.yaml"),
		DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter, log.FileWithLineNumCaller()),
	)
}

func initAliyunOSS() {
	aliyunoss.InitAliyunOSSManager(path.Join(env.ConfigDirPath(), "oss.yaml"))
}

func initGeeTest() {
	geetest.InitVerifier(geetest.LoadConfig(path.Join(env.ConfigDirPath(), "geetest.yaml")))
}

func initNATS(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	defaultNATS.SetManager(
		path.Join(env.ConfigDirPath(), "nats.yaml"),
		DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter, log.FileWithLineNumCaller()),
	)
}

func initHTTPClient(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter io.Writer) {
	resty.SetClient(DefaultCronRoutingHelper(debugWriter, infoWriter, warnErrorFatalWriter, cronDebugWriter, cronInfoWriter, cronWarnErrorFatalWriter, nil))
}
