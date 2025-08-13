package websocket

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	// 捕获系统信号以优雅地关闭调度器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	msgCh := pubSub.Channel(
		redis.WithChannelSize(10000),
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigChan:
			return errors.New(fmt.Sprintf("Received signal: %v, Subscribe shutdown", sig))
		case msg, ok := <-msgCh:
			if !ok {
				return errors.New("pubSub Channel closed")
			}
			ctx = context.WithValue(ctx, log.DefaultLogIdKey, utils.UniqueID())
			f(ctx, msg)
		}
	}
}

func (r *RedisPubSub) getRandomChannel() string {
	return r.channels[utils.RandIntN(len(r.channels))]
}

func (r *RedisPubSub) Publish(ctx context.Context, redisClient *redis.Client, msg any) *redis.IntCmd {
	return redisClient.Publish(ctx, r.getRandomChannel(), msg)
}
