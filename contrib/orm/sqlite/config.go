package sqlite

import "github.com/bpcoder16/Chestnut/v3/core/utils"

type Config struct {
	DSN          string
	MaxIdleConns int `json:"maxIdleConns"`
	MaxOpenConns int `json:"maxOpenConns"`
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load Sqlite conf err:" + err.Error())
	}
	return &config
}
