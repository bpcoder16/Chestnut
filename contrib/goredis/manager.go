package goredis

import (
	"context"
	"errors"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Manager struct {
	client *redis.Client
	logger *log.Helper
	config *Config
}

func NewManager(configPath string, logger *log.Helper) *Manager {
	manager := &Manager{
		logger: logger,
		config: loadConfig(configPath),
		client: nil,
	}
	manager.connect()
	return manager
}

func (m *Manager) Client() *redis.Client {
	return m.client
}

func (m *Manager) connect() {
	m.client = redis.NewClient(&redis.Options{
		Addr:         m.config.Host + ":" + strconv.Itoa(m.config.Port),
		Username:     m.config.Username,
		Password:     m.config.Password,
		DB:           m.config.DB,
		MaxRetries:   m.config.MaxRetries,
		DialTimeout:  m.config.DialTimeoutMillisecond * time.Millisecond,
		ReadTimeout:  m.config.ReadTimeoutMillisecond * time.Millisecond,
		WriteTimeout: m.config.WriteTimeoutMillisecond * time.Millisecond,
		PoolFIFO:     m.config.PoolFIFO,
		PoolSize:     m.config.PoolSize,
		//PoolTimeout:  200 * time.Millisecond,
		MinIdleConns:    m.config.MinIdleConns,
		MaxIdleConns:    m.config.MaxIdleConns,
		ConnMaxIdleTime: m.config.ConnMaxIdleTimeMinute * time.Minute,
		//ConnMaxLifetime: 2 * time.Hour,
	})
	m.client.AddHook(NewLoggerHook(m.logger))
	err := m.client.Get(context.Background(), "testConnect").Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		panic(m.config.Host + ":" + strconv.Itoa(m.config.Port) + ", failed to connect redis: " + err.Error())
	}
}
