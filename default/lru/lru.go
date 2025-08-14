package lru

import (
	"github.com/bpcoder16/Chestnut/v2/contrib/lru"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	gLRU "github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

var defaultManager *lru.Manager

func SetManager(configPath string, logger *log.Helper) {
	defaultManager = lru.NewManager(configPath, logger)
}

func DefaultLRUCache() *gLRU.Cache[string, any] {
	return defaultManager.LRUCache()
}

func DefaultExpireLRUCache() *expirable.LRU[string, any] {
	return defaultManager.ExpireLRUCache()
}
