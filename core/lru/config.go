package lru

import (
	"github.com/bpcoder16/Chestnut/core/utils"
	"github.com/bpcoder16/Chestnut/modules/appconfig/env"
)

func init() {
	loadConfig()
}

var config Config

func loadConfig() {
	err := utils.ParseJSONFile(env.RootPath()+"/conf/lru.json", &config)
	if err != nil {
		panic("load lru config err:" + err.Error())
	}
}
