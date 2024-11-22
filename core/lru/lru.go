package lru

import (
	"context"
	"encoding/json"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/bpcoder16/Chestnut/modules/appconfig/env"
	"github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"time"
)

var (
	DefaultLRUCache       *lru.Cache[string, any]
	DefaultExpireLRUCache *expirable.LRU[string, any]
)

func InitLRU() {
	var err error
	DefaultLRUCache, err = lru.NewWithEvict[string, any](config.Default.Size, defaultOnEvicted)
	if err != nil {
		panic("initLruCacheDefault.Err:" + err.Error())
	}
	DefaultExpireLRUCache = expirable.NewLRU[string, any](config.Expire.Size, defaultOnEvicted, time.Duration(config.Expire.TTLMillisecond)*time.Millisecond)
}

func defaultOnEvicted(key string, value any) {
	if env.RunMode() != env.RunModeRelease {
		OnEvictedValueJ, _ := json.Marshal(map[string]interface{}{
			"key":   key,
			"value": value,
		})
		logit.Context(context.WithValue(context.Background(), log.DefaultMessageKey, "LruCache")).DebugW("OnEvicted", string(OnEvictedValueJ))
	}
}
