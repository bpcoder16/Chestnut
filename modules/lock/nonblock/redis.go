package nonblock

import (
	"context"
	"fmt"
	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func RedisLock(ctx context.Context, redisClient *redis.Client, lockName string, deadLockExpireTime time.Duration) bool {
	timeNow := time.Now()
	cacheValue := strconv.Itoa(int(timeNow.Add(deadLockExpireTime).Unix()))
	success, err := redisClient.SetNX(ctx, lockName, cacheValue, deadLockExpireTime).Result()

	if err != nil {
		logit.Context(ctx).WarnW("RedisLockErr", err.Error())
		return false
	}

	// 防止死锁
	if !success {
		if expireTimeStr, errRedis := redisClient.Get(ctx, lockName).Result(); errRedis == nil {
			if expireTimeRedis, errStr := strconv.Atoi(expireTimeStr); errStr == nil {
				if timeNow.Unix() > int64(expireTimeRedis) {
					redisClient.Del(ctx, lockName)
				}
			} else {
				redisClient.Del(ctx, lockName)
			}
		} else {
			redisClient.Del(ctx, lockName)
		}
	}

	return success
}

func RedisUnlock(ctx context.Context, redisClient *redis.Client, lockName string) {
	redisClient.Del(ctx, lockName)
}

func BizBatchLock(ctx context.Context, redisClient *redis.Client, prefixLockName string, params ...any) (func(), bool) {
	return func() {
			bizBatchUnLock(ctx, redisClient, prefixLockName, params...)
		}, func() bool {
			unLockParams := make([]any, 0, len(params))
			for _, param := range params {
				var lockKey string
				switch v := param.(type) {
				case []any:
					// param 是数组，展开用作 sprintf
					lockKey = fmt.Sprintf(prefixLockName, v...)
				default:
					// param 是单值
					lockKey = fmt.Sprintf(prefixLockName, v)
				}
				success := RedisLock(ctx, redisClient, lockKey, time.Minute)
				if !success {
					bizBatchUnLock(ctx, redisClient, prefixLockName, unLockParams...)
					return false
				}
				unLockParams = append(unLockParams, param)
			}
			return true
		}()
}

func bizBatchUnLock(ctx context.Context, redisClient *redis.Client, prefixLockName string, params ...any) {
	for _, param := range params {
		var lockKey string
		switch v := param.(type) {
		case []any:
			// param 是数组，展开用作 sprintf
			lockKey = fmt.Sprintf(prefixLockName, v...)
		default:
			// param 是单值
			lockKey = fmt.Sprintf(prefixLockName, v)
		}
		if len(lockKey) > 0 {
			RedisUnlock(ctx, redisClient, lockKey)
		}
	}
}
