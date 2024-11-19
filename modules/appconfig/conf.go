package appconfig

import (
	"errors"
	"github.com/bpcoder16/Chestnut/core/utils"
	"github.com/bpcoder16/Chestnut/modules/appconfig/env"
	"path/filepath"
)

type AppConfig struct {
	ConfPath string

	Env env.Option

	FilterKeys   []string
	FilterValues []string

	DefaultMySQLSupport      bool
	DefaultClickhouseSupport bool
	DefaultRedisSupport      bool
	DefaultMongoDBSupport    bool
	AliyunOSSSupport         bool
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

func ParseConfig(path string, configPtr *AppConfig) (err error) {
	var confPath string
	confPath, err = filepath.Abs(path)
	if confPath, err = filepath.Abs(path); err == nil {
		if err = utils.ParseJSONFile(confPath, configPtr); err == nil {
			err = configPtr.Check()
		}
	}
	configPtr.ConfPath = confPath
	return
}
