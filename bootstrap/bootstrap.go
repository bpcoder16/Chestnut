package bootstrap

import (
	"context"
	"io"
	"path"
	"time"

	"github.com/bpcoder16/Chestnut/v2/appconfig"
	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"github.com/bpcoder16/Chestnut/v2/contrib/aliyun/oss"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/default/clickhouse"
	"github.com/bpcoder16/Chestnut/v2/default/lru"
	"github.com/bpcoder16/Chestnut/v2/default/mongodb"
	"github.com/bpcoder16/Chestnut/v2/default/mysql"
	"github.com/bpcoder16/Chestnut/v2/default/redis"
	"github.com/bpcoder16/Chestnut/v2/default/resty"
	"github.com/bpcoder16/Chestnut/v2/default/sqlite"
	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/bpcoder16/Chestnut/v2/modules/initdb"
	"github.com/bpcoder16/Chestnut/v2/modules/zaplogger"
)

func MustInit(ctx context.Context, config *appconfig.AppConfig, funcList ...func(ctx context.Context, debugWriter, infoWriter, warnErrorFatalWriter io.Writer)) {
	time.Local = env.TimeLocation()

	if config.Log.StdRedirectFileSupport {
		zaplogger.StdRedirectFile(config.Log.LogDir)
	}
	debugWriter, infoWriter, warnErrorFatalWriter := getWriters(config)

	initLoggers(ctx, config, debugWriter, infoWriter, warnErrorFatalWriter)

	initDefault(ctx, config, debugWriter, infoWriter, warnErrorFatalWriter)

	initHTTPClient(debugWriter, infoWriter, warnErrorFatalWriter)

	if config.MigrateService.Support &&
		config.MigrateService.DefaultVersion > 0 &&
		len(config.MigrateService.FileVersionSaveDir) > 0 &&
		len(config.MigrateService.MigrateSQLFileDir) > 0 {
		switch {
		case config.Default.MySQLSupport:
			initdb.Init(
				ctx,
				config.MigrateService.DefaultVersion,
				mysql.MasterDB(),
				config.MigrateService.FileVersionSaveDir,
				config.MigrateService.MigrateSQLFileDir,
			)
		case config.Default.SQLiteSupport:
			initdb.Init(
				ctx,
				config.MigrateService.DefaultVersion,
				sqlite.DefaultClient(),
				config.MigrateService.FileVersionSaveDir,
				config.MigrateService.MigrateSQLFileDir,
			)
		}
	}

	for _, fn := range funcList {
		fn(ctx, debugWriter, infoWriter, warnErrorFatalWriter)
	}
}

func getWriters(config *appconfig.AppConfig) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	if config.Log.UseRotateLog {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetFileRotateLogWriters(config.Log.LogDir, env.AppName(), env.AppName())
	} else {
		debugWriter, infoWriter, warnErrorFatalWriter = zaplogger.GetStandardWriters(config.Log.LogDir, env.AppName(), env.AppName())
	}
	return
}

func initLoggers(_ context.Context, config *appconfig.AppConfig, debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
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
}

func initDefault(ctx context.Context, config *appconfig.AppConfig, debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	if config.Default.MySQLSupport {
		initMySQL(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.Default.SQLiteSupport {
		initSQLite(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.Default.ClickhouseSupport {
		initClickhouse(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.Default.RedisSupport {
		initRedis(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.Default.MongoDBSupport {
		initMongoDB(ctx, debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.Default.LRUCacheSupport {
		initLRUCache(debugWriter, infoWriter, warnErrorFatalWriter)
	}
	if config.Default.AliyunOSSSupport {
		initAliyunOSS()
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

func initMySQL(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	mysql.SetManager(
		path.Join(env.ConfigDirPath(), "mysql.yaml"),
		DefaultHelper(debugWriter, infoWriter, warnErrorFatalWriter, nil),
	)
}

func initSQLite(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	sqlite.SetManager(
		path.Join(env.ConfigDirPath(), "sqlite.yaml"),
		DefaultHelper(debugWriter, infoWriter, warnErrorFatalWriter, nil),
	)
}

func initClickhouse(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	clickhouse.SetManager(
		path.Join(env.ConfigDirPath(), "clickhouse.yaml"),
		DefaultHelper(debugWriter, infoWriter, warnErrorFatalWriter, nil),
	)
}

func initRedis(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	redis.SetManager(
		path.Join(env.ConfigDirPath(), "redis.yaml"),
		DefaultHelper(debugWriter, infoWriter, warnErrorFatalWriter, log.FileWithLineNumCallerRedis()),
	)
}

func initMongoDB(ctx context.Context, debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	mongodb.SetManager(
		ctx,
		path.Join(env.ConfigDirPath(), "mongodb.yaml"),
		DefaultHelper(debugWriter, infoWriter, warnErrorFatalWriter, log.FileWithLineNumCaller()),
	)
}

func initLRUCache(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	lru.SetManager(
		path.Join(env.ConfigDirPath(), "lru.yaml"),
		DefaultHelper(debugWriter, infoWriter, warnErrorFatalWriter, log.FileWithLineNumCaller()),
	)
}

func initAliyunOSS() {
	oss.InitAliyunOSS(path.Join(env.ConfigDirPath(), "aliyun.yaml"))
}

func initHTTPClient(debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	resty.SetClient(DefaultHelper(debugWriter, infoWriter, warnErrorFatalWriter, nil))
}
