package goredis

import (
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"time"
)

type Config struct {
	Host                    string        `json:"host"`
	Port                    int           `json:"port"`
	DB                      int           `json:"db"`
	Username                string        `json:"username"`
	Password                string        `json:"password"`
	MaxRetries              int           `json:"maxRetries"`
	DialTimeoutMillisecond  time.Duration `json:"dialTimeoutMillisecond"`
	ReadTimeoutMillisecond  time.Duration `json:"readTimeoutMillisecond"`
	WriteTimeoutMillisecond time.Duration `json:"writeTimeoutMillisecond"`
	PoolFIFO                bool          `json:"poolFIFO"`
	PoolSize                int           `json:"poolSize"`
	MinIdleConns            int           `json:"minIdleConns"`
	MaxIdleConns            int           `json:"maxIdleConns"`
	ConnMaxIdleTimeMinute   time.Duration `json:"connMaxIdleTimeMinute"`
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load Redis conf err:" + err.Error())
	}
	return &config
}
