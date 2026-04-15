package cronserver

import (
	"github.com/bpcoder16/Chestnut/v4/core/utils"
)

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load cron conf err:" + err.Error())
	}
	return &config
}
