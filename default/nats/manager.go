package nats

import (
	"github.com/bpcoder16/Chestnut/v4/contrib/gonats"
	"github.com/bpcoder16/Chestnut/v4/core/cdefer"
	"github.com/bpcoder16/Chestnut/v4/core/log"
)

var defaultManager *gonats.Manager

// SetManager 初始化全局默认 NATS Manager，由 bootstrap 在启动时调用。
// 会自动将 Close() 注册到 cdefer，应用退出时优雅关闭连接。
func SetManager(configPath string, logger *log.Helper) {
	defaultManager = gonats.NewManager(configPath, logger)
	cdefer.RegisterDeferFunc(defaultManager.Close)
}

// DefaultManager 返回全局默认 Manager，用于访问 Core NATS 和 JetStream 能力。
func DefaultManager() *gonats.Manager {
	return defaultManager
}
