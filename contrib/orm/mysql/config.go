package mysql

import "github.com/bpcoder16/Chestnut/v4/core/utils"

type ConfigItem struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Database     string `json:"database"`
	Charset      string `json:"charset"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	MaxIdleConns int    `json:"maxIdleConns"`
	MaxOpenConns int    `json:"maxOpenConns"`
}

type Config struct {
	Master *ConfigItem   `json:"master"`
	Slaves []*ConfigItem `json:"slaves"`
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load MySQL conf err:" + err.Error())
	}
	return &config
}
