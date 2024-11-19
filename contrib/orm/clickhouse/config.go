package clickhouse

import "github.com/bpcoder16/Chestnut/core/utils"

type ConfigItem struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Database     string `json:"database"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	MaxIdleConns int    `json:"maxIdleConns"`
	MaxOpenConns int    `json:"maxOpenConns"`
}

type Config struct {
	Master *ConfigItem   `json:"master"`
	Slaves []*ConfigItem `json:"slaves"`
}

func loadClickhouseConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load Clickhouse conf err:" + err.Error())
	}
	return &config
}
