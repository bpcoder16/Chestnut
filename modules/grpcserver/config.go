package grpcserver

import (
	"github.com/bpcoder16/Chestnut/v2/core/utils"
)

type Config struct {
	Port string
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load grpc Server conf err:" + err.Error())
	}
	return &config
}
