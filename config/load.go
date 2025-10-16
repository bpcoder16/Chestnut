package config

import (
	"flag"
	"fmt"
	"path"
	"path/filepath"

	"github.com/bpcoder16/Chestnut/v3/config/env"
	"github.com/bpcoder16/Chestnut/v3/contrib/nacos"
	"github.com/bpcoder16/Chestnut/v3/core/utils"
)

func MustLoadBaseConfig(confPath string) *AppConfig {
	cfgMode := flag.String(
		"cfg-mode",
		BaseConfigModeLocal,
		"配置文件方案: local(本地文件) 或 nacos(Nacos配置中心)",
	)

	flag.Parse()

	var baseConfig Base
	var err error
	confPath, err = filepath.Abs(path.Join(utils.RootPath(), confPath))
	if err != nil {
		panic("load BaseConfig err:" + err.Error())
	}
	err = utils.ParseFile(confPath, &baseConfig)
	if err != nil {
		panic("Parse BaseConfig err:" + err.Error())
	}

	var appConfig AppConfig
	switch *cfgMode {
	case BaseConfigModeLocal:
		appConfig = loadLocalAppConfig(baseConfig.Local)
	case BaseConfigModeNacos:
		appConfig = loadNacosAppConfig(baseConfig.Nacos)
	default:
		panic(fmt.Sprintf("不支持的配置模式: %s, 仅支持 'local' 或 'nacos'", *cfgMode))
	}

	return &appConfig
}

func loadLocalAppConfig(baseConfig BaseLocal) AppConfig {
	var config AppConfig
	err := ParseLocalConfig(
		path.Join(utils.RootPath(), baseConfig.ConfigDirName, baseConfig.ConfigFile),
		&config,
	)
	if err != nil {
		panic("parse local app config failed: " + err.Error())
	}
	env.Default = env.New(config.Env)
	return config
}

func loadNacosAppConfig(baseConfig BaseNacos) AppConfig {
	nacos.Init(baseConfig.ServerConfigs, baseConfig.ClientConfig)
	var config AppConfig
	err := ParseNacosConfig(baseConfig.ConfigFile, baseConfig.Group, &config)
	if err != nil {
		panic("parse nacos app config failed: " + err.Error())
	}
	env.Default = env.New(config.Env)
	return config
}
