package gonats

import (
	"context"
	"fmt"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// LogIdHeader 是 NATS 消息 Header 中传递 Log-Id 的键名。
// 发布方自动将 ctx 中的 logId 写入此 Header，消费方从此 Header 还原 logId 到新 ctx。
const LogIdHeader = "Nats-Log-Id"

// injectLogId 将 ctx 中的 logId 写入 msg Header。
// 若 ctx 中不存在 logId，则生成一个新的 ID 写入。
func injectLogId(ctx context.Context, msg *nats.Msg) {
	logId := ctx.Value(log.DefaultLogIdKey)
	if logId == nil || logId == "" {
		logId = utils.UniqueID()
	}
	if msg.Header == nil {
		msg.Header = nats.Header{}
	}
	msg.Header.Set(LogIdHeader, fmt.Sprintf("%v", logId))
}

// extractLogId 从 msg Header 中提取 logId 并注入到 ctx 中返回。
// 若 Header 中不存在 logId，则生成一个新的 ID。
// 同时设置 DefaultMessageKey 为 "NATS"，便于日志中识别消息来源。
func extractLogId(ctx context.Context, msg *nats.Msg) context.Context {
	return extractLogIdFromHeader(ctx, msg.Header)
}

// extractLogIdFromJSMsg 从 jetstream.Msg Header 中提取 logId 并注入到 ctx 中返回。
func extractLogIdFromJSMsg(ctx context.Context, msg jetstream.Msg) context.Context {
	return extractLogIdFromHeader(ctx, msg.Headers())
}

func extractLogIdFromHeader(ctx context.Context, header nats.Header) context.Context {
	logId := header.Get(LogIdHeader)
	if logId == "" {
		logId = utils.UniqueID()
	}
	ctx = context.WithValue(ctx, log.DefaultMessageKey, "NATS")
	return context.WithValue(ctx, log.DefaultLogIdKey, logId)
}
