package goredis

import (
	"context"
	"net"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/redis/go-redis/v9"
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
		//begin := time.Now()
		conn, err = next(ctx, network, addr)
		//elapsed := time.Since(begin)
		//fmt.Printf("redis cmd[connect %s] costTime[%s]\n", addr, fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6))
		return
	}
}

// sensitiveRedisCommands 包含可能携带密码的命令名（小写），日志中需脱敏。
var sensitiveRedisCommands = map[string]struct{}{
	"hello": {},
	"auth":  {},
}

// internalRedisCommands 是连接握手或健康探测时发送的内部命令，无业务价值，不记录日志。
var internalRedisCommands = map[string]struct{}{
	"client": {},
	"ping":   {},
}

func redactCmd(cmd redis.Cmder) (string, bool) {
	if _, ok := internalRedisCommands[cmd.Name()]; ok {
		return "", false
	}
	if _, ok := sensitiveRedisCommands[cmd.Name()]; ok {
		return cmd.Name() + " [redacted]", true
	}
	return cmd.String(), true
}

func (l *LoggerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) (err error) {
		ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "Redis")
		begin := time.Now()
		err = next(ctx, cmd)
		elapsed := time.Since(begin)
		if cmdStr, ok := redactCmd(cmd); ok {
			l.Helper.WithContext(ctx).DebugW(
				"cmd", cmdStr,
				"costTime", utils.ShowDurationString(elapsed),
			)
		}
		return
	}
}

func (l *LoggerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmdList []redis.Cmder) error {
		return next(ctx, cmdList)
	}
}
