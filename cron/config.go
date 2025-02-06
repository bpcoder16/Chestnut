package cron

import (
	"github.com/bpcoder16/Chestnut/v2/core/utils"
)

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load cron conf err:" + err.Error())
	}
	return &config
}
