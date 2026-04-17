package goredis

import (
	"context"
	"strconv"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/redis/go-redis/v9"
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

// Close 关闭 Redis 连接池，释放所有连接。
// 应在应用退出时调用（如注册到 cdefer）。
func (m *Manager) Close() {
	if err := m.client.Close(); err != nil {
		m.logger.WarnW("Redis.Close", "failed to close connection pool", "err", err)
	}
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
	if err := m.client.Ping(context.Background()).Err(); err != nil {
		panic(m.config.Host + ":" + strconv.Itoa(m.config.Port) + ", failed to connect redis: " + err.Error())
	}
}
