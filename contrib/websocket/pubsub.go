package websocket

import (
	"context"
	"github.com/bpcoder16/Chestnut/core/utils"
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

func (r *RedisPubSub) Subscribe(ctx context.Context, redisClient *redis.Client, f func(*redis.Message)) error {
	pubSub := redisClient.Subscribe(ctx, r.channels...)
	defer func() {
		_ = pubSub.Close()
	}()
	for {
		msg, errR := pubSub.ReceiveMessage(ctx)
		if errR != nil {
			return errR
		}
		f(msg)
	}
}

func (r *RedisPubSub) getRandomChannel() string {
	return r.channels[utils.RandIntN(len(r.channels))]
}

func (r *RedisPubSub) Publish(ctx context.Context, redisClient *redis.Client, msg any) *redis.IntCmd {
	return redisClient.Publish(ctx, r.getRandomChannel(), msg)
}
