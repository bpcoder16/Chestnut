package websocket

import (
	"context"
	"errors"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/redis/go-redis/v9"
)

type RedisPubSub struct {
	channels []string
}

func NewRedisPubSub(channels ...string) (*RedisPubSub, error) {
	if len(channels) == 0 {
		return nil, errors.New("redis pubsub channels empty")
	}
	for _, channel := range channels {
		if channel == "" {
			return nil, errors.New("redis pubsub channel empty")
		}
	}
	return &RedisPubSub{
		channels: channels,
	}, nil
}

func (r *RedisPubSub) Subscribe(ctx context.Context, redisClient *redis.Client, f func(context.Context, *redis.Message)) error {
	if r == nil || len(r.channels) == 0 {
		return errors.New("redis pubsub channels empty")
	}
	if redisClient == nil {
		return errors.New("redis client nil")
	}
	if f == nil {
		return errors.New("redis pubsub handler nil")
	}
	ctx = context.WithValue(ctx, log.DefaultMessageKey, "RedisPubSub.Subscribe")
	pubSub := redisClient.Subscribe(ctx, r.channels...)
	defer func() {
		_ = pubSub.Close()
	}()

	msgCh := pubSub.Channel(
		redis.WithChannelSize(10000),
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgCh:
			if !ok {
				return errors.New("pubSub Channel closed")
			}
			ctx = context.WithValue(ctx, log.DefaultLogIdKey, utils.UniqueID())
			f(ctx, msg)
		}
	}
}

func (r *RedisPubSub) getRandomChannel() (string, error) {
	if r == nil || len(r.channels) == 0 {
		return "", errors.New("redis pubsub channels empty")
	}
	if len(r.channels) == 1 {
		return r.channels[0], nil
	}
	return r.channels[utils.RandIntN(len(r.channels))], nil
}

func (r *RedisPubSub) Publish(ctx context.Context, redisClient *redis.Client, msg any) (*redis.IntCmd, error) {
	if redisClient == nil {
		return nil, errors.New("redis client nil")
	}
	channel, err := r.getRandomChannel()
	if err != nil {
		return nil, err
	}
	return redisClient.Publish(ctx, channel, msg), nil
}
