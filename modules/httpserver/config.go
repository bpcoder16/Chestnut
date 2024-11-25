package httpserver

import "github.com/bpcoder16/Chestnut/core/utils"

type Config struct {
	Server struct {
		Port string `json:"port"`
	}
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load HTTP Server conf err:" + err.Error())
	}
	return &config
}
