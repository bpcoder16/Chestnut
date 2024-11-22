package lru

import (
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"time"
)

type Manager struct {
	lruCache       *lru.Cache[string, any]
	expireLRUCache *expirable.LRU[string, any]
	logger         *log.Helper
	config         *Config
}

func NewManager(configPath string, logger *log.Helper) *Manager {
	manager := &Manager{
		lruCache:       nil,
		expireLRUCache: nil,
		logger:         logger,
		config:         loadConfig(configPath),
	}
	manager.initLRUCache()
	return manager
}

func (m *Manager) initLRUCache() {
	var err error
	m.lruCache, err = lru.NewWithEvict[string, any](m.config.Default.Size, defaultOnEvictedFunc(m.logger))
	if err != nil {
		panic("initLRUCacheDefault.Err:" + err.Error())
	}
	m.expireLRUCache = expirable.NewLRU[string, any](m.config.Expire.Size, defaultOnEvictedFunc(m.logger), time.Duration(m.config.Expire.TTLMillisecond)*time.Millisecond)
}

func (m *Manager) LRUCache() *lru.Cache[string, any] {
	return m.lruCache
}

func (m *Manager) ExpireLRUCache() *expirable.LRU[string, any] {
	return m.expireLRUCache
}
