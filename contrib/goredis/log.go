package goredis

import (
	"context"
	"fmt"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/redis/go-redis/v9"
	"net"
	"time"
)

type LoggerHook struct {
	*log.Helper
}

func NewLoggerHook(helper *log.Helper) *LoggerHook {
	return &LoggerHook{
		Helper: helper,
	}
}

func (l *LoggerHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
		ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "Redis")
		begin := time.Now()
		conn, err = next(ctx, network, addr)
		elapsed := time.Since(begin)
		fmt.Printf("redis cmd[connect %s] costTime[%s]\n", addr, fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6))
		return
	}
}

func (l *LoggerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) (err error) {
		ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "Redis")
		begin := time.Now()
		err = next(ctx, cmd)
		elapsed := time.Since(begin)
		l.Helper.WithContext(ctx).DebugW(
			"cmd", cmd.String(),
			"costTime", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
		)
		return
	}
}

func (l *LoggerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmdList []redis.Cmder) error {
		return next(ctx, cmdList)
	}
}
