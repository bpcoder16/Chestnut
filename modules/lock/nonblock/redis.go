package nonblock

import (
	"context"
	"github.com/bpcoder16/Chestnut/logit"
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
