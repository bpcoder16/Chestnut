package goredis

import (
	"context"
	"errors"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type RedisManager struct {
	client *redis.Client
	logger *log.Helper
	config *Config
}

func NewRedisManager(configPath string, logger *log.Helper) *RedisManager {
	manager := &RedisManager{
		logger: logger,
		config: loadRedisConfig(configPath),
		client: nil,
	}
	manager.connect()
	return manager
}

func (r *RedisManager) Client() *redis.Client {
	return r.client
}

func (r *RedisManager) connect() {
	r.client = redis.NewClient(&redis.Options{
		Addr:         r.config.Host + ":" + strconv.Itoa(r.config.Port),
		Username:     r.config.Username,
		Password:     r.config.Password,
		DB:           r.config.DB,
		MaxRetries:   r.config.MaxRetries,
		DialTimeout:  200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
		WriteTimeout: 200 * time.Millisecond,
		PoolFIFO:     true,
		PoolSize:     100,
		//PoolTimeout:  200 * time.Millisecond,
		MinIdleConns:    20,
		MaxIdleConns:    50,
		ConnMaxIdleTime: 10 * time.Minute,
		//ConnMaxLifetime: 2 * time.Hour,
	})
	r.client.AddHook(NewLoggerHook(r.logger))
	err := r.client.Get(context.Background(), "testConnect").Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		panic(r.config.Host + ":" + strconv.Itoa(r.config.Port) + ", failed to connect redis: " + err.Error())
	}
}
