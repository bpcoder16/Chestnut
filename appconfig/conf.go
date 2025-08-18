package appconfig

import (
	"errors"
	"path"
	"path/filepath"

	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
)

type AppConfig struct {
	Env env.Option

	FilterKeys   []string
	Log          Log
	Default      Default
	AsyncService AsyncService
}

type Log struct {
	LogDir                 string
	UseRotateLog           bool
	StdRedirectFileSupport bool
}

type Default struct {
	MySQLSupport         bool
	SQLiteSupport        bool
	ClickhouseSupport    bool
	RedisSupport         bool
	MongoDBSupport       bool
	LRUCacheSupport      bool
	AliyunOSSSupport     bool
	LocalLockPoolSupport bool
}

type AsyncService struct {
	Support         bool
	QueueSize       int
	ConsumerSize    int
	TaskMaxRetryCnt int
}

func (c *AppConfig) Check() (err error) {
	if len(c.Env.AppName) == 0 {
		err = errors.New("AppName required")
	}
	switch c.Env.RunMode {
	case env.RunModeDebug, env.RunModeTest, env.RunModeRelease:
	default:
		err = errors.New("invalid runMode: " + c.Env.RunMode)
	}
	return err
}

func ParseConfig(confPath string, configPtr *AppConfig) (err error) {
	if confPath, err = filepath.Abs(confPath); err == nil {
		if err = utils.ParseFile(confPath, configPtr); err == nil {
			err = configPtr.Check()
		}
	}
	if len(configPtr.Log.LogDir) == 0 {
		configPtr.Log.LogDir = path.Join(utils.RootPath(), "log")
	}
	return
}
