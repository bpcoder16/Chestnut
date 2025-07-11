package appconfig

import (
	"errors"
	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"path"
	"path/filepath"
)

type AppConfig struct {
	Env env.Option

	FilterKeys []string

	LogDir          string
	NotUseRotateLog bool

	StdRedirectFileSupport   bool
	DefaultMySQLSupport      bool
	DefaultClickhouseSupport bool
	DefaultRedisSupport      bool
	DefaultMongoDBSupport    bool
	AliyunOSSSupport         bool
	UseLRUCache              bool
	QueueSize                int
	ConsumerSize             int
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
		if err = utils.ParseJSONFile(confPath, configPtr); err == nil {
			err = configPtr.Check()
		}
	}
	if len(configPtr.LogDir) == 0 {
		configPtr.LogDir = path.Join(utils.RootPath(), "log")
	}
	return
}
