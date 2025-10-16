package config

import (
	"errors"

	"github.com/bpcoder16/Chestnut/v3/config/env"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
)

const (
	BaseConfigModeLocal = "local"

	BaseConfigModeNacos = "nacos"
)

type BaseLocal struct {
	ConfigDirName string
	ConfigFile    string
}

type BaseNacos struct {
	ServerConfigs []constant.ServerConfig
	ClientConfig  *constant.ClientConfig
	Group         string
	ConfigFile    string
}

type Base struct {
	Local BaseLocal
	Nacos BaseNacos
}

type AppConfig struct {
	Env env.Option
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
