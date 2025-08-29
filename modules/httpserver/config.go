package httpserver

import (
	"time"

	"github.com/bpcoder16/Chestnut/v2/core/utils"
)

type Config struct {
	Port                         string
	ReadTimeoutMillisecond       time.Duration // 读取数据最大时间
	ReadHeaderTimeoutMillisecond time.Duration // 读取请求头最大时间
	WriteTimeoutMillisecond      time.Duration // 写响应最大时间
	IdleTimeoutMillisecond       time.Duration // 空闲连接最大时间
	MaxHeaderBytes               int           // 最大请求头大小
	IsOpenConnStateTraceLog      bool          // 是否开启连接状态追踪日志
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load HTTP Server conf err:" + err.Error())
	}
	return &config
}
