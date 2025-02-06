package lru

import (
	"github.com/bpcoder16/Chestnut/v2/core/utils"
)

type Config struct {
	Default struct {
		Size int `json:"size"`
	} `json:"default"`
	Expire struct {
		Size           int   `json:"size"`
		TTLMillisecond int64 `json:"ttlMillisecond"`
	} `json:"expire"`
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load LRU conf err:" + err.Error())
	}
	return &config
}
