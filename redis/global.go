package redis

import (
	"github.com/bpcoder16/Chestnut/v2/contrib/goredis"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/redis/go-redis/v9"
)

var defaultManager *goredis.Manager

func SetManager(configPath string, logger *log.Helper) {
	defaultManager = goredis.NewManager(configPath, logger)
}

func DefaultClient() *redis.Client {
	return defaultManager.Client()
}
