package log

import "context"

const (
	// MicroServiceLogIdHeader 是微服务间 HTTP 调用传递 logId 的 Header 名。
	MicroServiceLogIdHeader = "MicroService-Log-Id"
)

// WithLogId 将 logId 注入 context，用于串联同一次请求的日志。
func WithLogId(ctx context.Context, logId string) context.Context {
	return context.WithValue(ctx, DefaultLogIdKey, logId)
}

// LogIdFromContext 从 context 中读取 logId，缺失时返回空字符串。
func LogIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	logId, _ := ctx.Value(DefaultLogIdKey).(string)
	return logId
}
