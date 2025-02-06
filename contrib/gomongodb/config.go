package gomongodb

import "github.com/bpcoder16/Chestnut/v2/core/utils"

type Config struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Database           string `json:"database"`
	MaxPoolSize        uint64 `json:"maxPoolSize"`
	MinPoolSize        uint64 `json:"minPoolSize"`
	MaxConnIdleTimeSec int    `json:"maxConnIdleTimeSec"`
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load MongoDB conf err:" + err.Error())
	}
	return &config
}
