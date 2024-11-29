package websocket

import (
	"context"
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
