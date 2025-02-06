package lru

import (
	"context"
	"encoding/json"
	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"github.com/bpcoder16/Chestnut/v2/core/log"
)

func defaultOnEvictedFunc(l *log.Helper) func(key string, value any) {
	return func(key string, value any) {
		if env.RunMode() != env.RunModeRelease {
			OnEvictedValueJ, _ := json.Marshal(map[string]interface{}{
				"key":   key,
				"value": value,
			})
			l.WithContext(context.WithValue(context.Background(), log.DefaultMessageKey, "LRUCache")).DebugW("OnEvicted", string(OnEvictedValueJ))
		}
	}
}
