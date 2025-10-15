package httpserver

import (
	"time"

	"github.com/bpcoder16/Chestnut/v3/core/utils"
)

type Config struct {
	Port                         string        // 端口号
	ReadTimeoutMillisecond       time.Duration // 读取数据最大时间
	ReadHeaderTimeoutMillisecond time.Duration // 读取请求头最大时间
	WriteTimeoutMillisecond      time.Duration // 写响应最大时间
	IdleTimeoutMillisecond       time.Duration // 空闲连接最大时间
	MaxHeaderBytes               int           // 最大请求头大小
	IsOpenConnStateTraceLog      bool          // 是否开启连接状态追踪日志
	ShutdownTimeoutSecond        int           // 优雅关闭超时时间（秒）
}

// loadConfig 加载HTTP服务器配置文件
// configPath: 配置文件路径
// 如果配置加载失败会panic，请确保配置文件存在且格式正确
func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load HTTP Server config failed: " + err.Error())
	}
	return &config
}
