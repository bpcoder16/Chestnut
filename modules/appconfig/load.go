package appconfig

import (
	"github.com/bpcoder16/Chestnut/core/utils"
	"github.com/bpcoder16/Chestnut/modules/appconfig/env"
)

// 配置文件使用 json 格式
// 配置文件强制路径为 根目录的 conf/app.json
// 并完成 env 的配置

// MustLoadAppConfig 加载 app.toml ,若失败，会 panic
func MustLoadAppConfig() *AppConfig {
	var config AppConfig
	err := ParseConfig(utils.RootPath()+"/conf/app.json", &config)
	if err != nil {
		panic("parse app config failed: " + err.Error())
	}
	env.Default = env.New(config.Env)
	return &config
}
