package lru

import (
	"context"
	"encoding/json"
	env2 "github.com/bpcoder16/Chestnut/appconfig/env"
	"github.com/bpcoder16/Chestnut/core/log"
)

func defaultOnEvictedFunc(l *log.Helper) func(key string, value any) {
	return func(key string, value any) {
		if env2.RunMode() != env2.RunModeRelease {
			OnEvictedValueJ, _ := json.Marshal(map[string]interface{}{
				"key":   key,
				"value": value,
			})
			l.WithContext(context.WithValue(context.Background(), log.DefaultMessageKey, "LRUCache")).DebugW("OnEvicted", string(OnEvictedValueJ))
		}
	}
}
