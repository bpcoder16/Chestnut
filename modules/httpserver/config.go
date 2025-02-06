package httpserver

import (
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"time"
)

type Config struct {
	Port                         string        `json:"port"`
	ReadTimeoutMillisecond       time.Duration `json:"readTimeoutMillisecond"`       // 读取数据最大时间
	ReadHeaderTimeoutMillisecond time.Duration `json:"readHeaderTimeoutMillisecond"` // 读取请求头最大时间
	WriteTimeoutMillisecond      time.Duration `json:"writeTimeoutMillisecond"`      // 写响应最大时间
	IdleTimeoutMillisecond       time.Duration `json:"idleTimeoutMillisecond"`       // 空闲连接最大时间
	MaxHeaderBytes               int           `json:"maxHeaderBytes"`               // 最大请求头大小
	IsOpenConnStateTraceLog      bool          `json:"isOpenConnStateTraceLog"`      // 是否开启连接状态追踪日志
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load HTTP Server conf err:" + err.Error())
	}
	return &config
}
