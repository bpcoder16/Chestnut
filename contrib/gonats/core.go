package gonats

import (
	"context"

	"github.com/nats-io/nats.go"
)

// Publish 向指定 Subject 发布消息（fire-and-forget，At-most-once）。
// 自动将 ctx 中的 logId 写入消息 Header（Nats-Log-Id），实现跨服务日志追踪。
//
// header 参数可选：不传则使用空 Header，传入则在其基础上追加 Nats-Log-Id。
// 消费方使用 Subscribe / QueueSubscribe 可自动还原 logId 到 ctx。
//
// 使用示例：
//
//	// 不带自定义 Header
//	nm.Publish(ctx, "orders.new", payload)
//
//	// 带自定义 Header（框架会额外追加 Nats-Log-Id）
//	nm.Publish(ctx, "orders.new", payload, nats.Header{"X-Source": []string{"api"}})
func (m *Manager) Publish(ctx context.Context, subject string, data []byte, headers ...nats.Header) error {
	var header nats.Header
	if len(headers) > 0 && headers[0] != nil {
		header = headers[0]
	} else {
		header = nats.Header{}
	}
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  header,
	}
	injectLogId(ctx, msg)
	if err := m.nc.PublishMsg(msg); err != nil {
		return err
	}
	m.logger.WithContext(ctx).DebugW("NATS.Publish", "message published", "subject", subject, "data", string(data), "headers", msg.Header)
	return nil
}

// Request 发送请求并等待响应，超时由 ctx 的 Deadline 控制（同步 RPC 模式）。
// 自动将 ctx 中的 logId 写入请求 Header，响应方通过 QueueSubscribe 接收时可获得同一 logId，
// 实现 RPC 调用链的完整日志追踪。响应方需调用 msg.Respond() 回复。
func (m *Manager) Request(ctx context.Context, subject string, data []byte) (*nats.Msg, error) {
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
	}
	injectLogId(ctx, msg)
	m.logger.WithContext(ctx).DebugW("NATS.Request", "sending request", "subject", subject, "data", string(data), "headers", msg.Header)
	reply, err := m.nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		m.logger.WithContext(ctx).ErrorW("NATS.Request", "request failed", "subject", subject, "err", err)
		return nil, err
	}
	m.logger.WithContext(ctx).DebugW("NATS.Request", "received reply", "subject", subject, "data", string(reply.Data), "headers", reply.Header)
	return reply, nil
}

// Subscribe 订阅指定 Subject，消息广播给所有订阅者（Pub/Sub 模式）。
// 支持通配符：* 匹配单 token，> 匹配多 token（如 "orders.>"）。
// 自动从消息 Header 还原 logId 并注入 ctx，与 Publish 配合实现跨服务日志追踪。
// 返回的 *nats.Subscription 可调用 Unsubscribe() / Drain() 取消订阅。
func (m *Manager) Subscribe(subject string, handler func(context.Context, *nats.Msg)) (*nats.Subscription, error) {
	return m.nc.Subscribe(subject, func(msg *nats.Msg) {
		msgCtx := extractLogId(context.Background(), msg)
		m.logger.WithContext(msgCtx).DebugW("NATS.Receive", "received message", "subject", msg.Subject, "data", string(msg.Data), "headers", msg.Header)
		handler(msgCtx, msg)
	})
}

// QueueSubscribe 队列订阅，同一队列组内只有一个成员收到消息（竞争消费/负载均衡）。
// 多个实例使用相同的 queue 名即可自动形成队列组，无需额外配置。
// 自动从消息 Header 还原 logId 并注入 ctx，与 Publish 配合实现跨服务日志追踪。
// 返回的 *nats.Subscription 可调用 Unsubscribe() / Drain() 取消订阅。
func (m *Manager) QueueSubscribe(subject, queue string, handler func(context.Context, *nats.Msg)) (*nats.Subscription, error) {
	return m.nc.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		msgCtx := extractLogId(context.Background(), msg)
		m.logger.WithContext(msgCtx).DebugW("NATS.Receive", "received message", "subject", msg.Subject, "queue", queue, "data", string(msg.Data), "headers", msg.Header)
		handler(msgCtx, msg)
	})
}
