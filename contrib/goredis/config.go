package goredis

import "github.com/bpcoder16/Chestnut/core/utils"

type Config struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	DB         int    `json:"db"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	MaxRetries int    `json:"maxRetries"`
}

func loadRedisConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load Redis conf err:" + err.Error())
	}
	return &config
}
