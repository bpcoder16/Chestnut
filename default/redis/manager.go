package redis

import (
	"github.com/bpcoder16/Chestnut/v4/contrib/goredis"
	"github.com/bpcoder16/Chestnut/v4/core/cdefer"
	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/redis/go-redis/v9"
)

var defaultManager *goredis.Manager

func SetManager(configPath string, logger *log.Helper) {
	defaultManager = goredis.NewManager(configPath, logger)
	cdefer.RegisterDeferFunc(defaultManager.Close)
}

func DefaultClient() *redis.Client {
	return defaultManager.Client()
}
