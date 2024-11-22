package nonblock

import (
	"context"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/bpcoder16/Chestnut/redis"
	"strconv"
	"time"
)

func RedisLock(ctx context.Context, lockName string, deadLockExpireTime time.Duration) bool {
	timeNow := time.Now()
	cacheValue := strconv.Itoa(int(timeNow.Add(deadLockExpireTime).Unix()))
	success, err := redis.DefaultClient().SetNX(ctx, lockName, cacheValue, deadLockExpireTime).Result()

	if err != nil {
		logit.Context(ctx).WarnW("RedisLockErr", err.Error())
		return false
	}

	// 防止死锁
	if !success {
		if expireTimeStr, errRedis := redis.DefaultClient().Get(ctx, lockName).Result(); errRedis == nil {
			if expireTimeRedis, errStr := strconv.Atoi(expireTimeStr); errStr == nil {
				if timeNow.Unix() > int64(expireTimeRedis) {
					redis.DefaultClient().Del(ctx, lockName)
				}
			} else {
				redis.DefaultClient().Del(ctx, lockName)
			}
		} else {
			redis.DefaultClient().Del(ctx, lockName)
		}
	}

	return success
}

func RedisUnlock(ctx context.Context, lockName string) {
	redis.DefaultClient().Del(ctx, lockName)
}
