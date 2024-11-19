package redis

import (
	"github.com/bpcoder16/Chestnut/contrib/goredis"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/redis/go-redis/v9"
)

var defaultRedisManager *goredis.RedisManager

func SetManager(configPath string, logger *log.Helper) {
	defaultRedisManager = goredis.NewRedisManager(configPath, logger)
}

func DefaultClient() *redis.Client {
	return defaultRedisManager.Client()
}
