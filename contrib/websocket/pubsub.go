package websocket

import (
	"context"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"github.com/redis/go-redis/v9"
)

type RedisPubSub struct {
	channels []string
}

func NewRedisPubSub(channels ...string) *RedisPubSub {
	return &RedisPubSub{
		channels: channels,
	}
}

func (r *RedisPubSub) Subscribe(ctx context.Context, redisClient *redis.Client, f func(context.Context, *redis.Message)) error {
	ctx = context.WithValue(ctx, log.DefaultMessageKey, "RedisPubSub.Subscribe")
	pubSub := redisClient.Subscribe(ctx, r.channels...)
	defer func() {
		_ = pubSub.Close()
	}()
	for {
		msg, errR := pubSub.ReceiveMessage(ctx)
		if errR != nil {
			return errR
		}
		ctx = context.WithValue(ctx, log.DefaultLogIdKey, utils.UniqueID())
		f(ctx, msg)
	}
}

func (r *RedisPubSub) getRandomChannel() string {
	return r.channels[utils.RandIntN(len(r.channels))]
}

func (r *RedisPubSub) Publish(ctx context.Context, redisClient *redis.Client, msg any) *redis.IntCmd {
	return redisClient.Publish(ctx, r.getRandomChannel(), msg)
}
