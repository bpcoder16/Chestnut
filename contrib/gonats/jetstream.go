package gonats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"
)

// MustRegisterStream 创建或更新 Stream 并保存到全局注册表，供后续通过 GetStream 取用。
// 失败时直接 panic，用于应用启动阶段初始化。
func (m *Manager) MustRegisterStream(ctx context.Context, cfg StreamConfig) jetstream.Stream {
	stream, err := m.js.CreateOrUpdateStream(ctx, buildStreamConfig(cfg))
	if err != nil {
		panic("failed to register NATS stream [" + cfg.Name + "]: " + err.Error())
	}
	m.streams.Store(cfg.Name, stream)
	return stream
}

// GetStream 从注册表中取出已注册的 Stream。
// 若 Stream 尚未通过 MustRegisterStream 注册，返回 error。
func (m *Manager) GetStream(name string) (jetstream.Stream, error) {
	v, ok := m.streams.Load(name)
	if !ok {
		return nil, fmt.Errorf("stream %q not registered, call MustRegisterStream first", name)
	}
	return v.(jetstream.Stream), nil
}

// DeleteStream 从服务端删除指定 Stream 及其所有消息和 Consumer，同时清除本地注册。
// 谨慎使用，数据不可恢复。
func (m *Manager) DeleteStream(ctx context.Context, name string) error {
	if err := m.js.DeleteStream(ctx, name); err != nil {
		return err
	}
	m.streams.Delete(name)
	return nil
}

// JSPublish 向 JetStream 同步发布消息，等待服务端 PubAck 确认持久化后返回。
// 返回的 PubAck 包含消息落盘的 Stream 名称和序列号。
// 自动将 ctx 中的 logId 注入消息 Header，实现跨服务日志追踪。
//
// header 可选：不传则使用空 Header，传入则在其基础上追加 Nats-Log-Id。
func (m *Manager) JSPublish(ctx context.Context, subject string, data []byte, headers ...nats.Header) (*jetstream.PubAck, error) {
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
	ack, err := m.js.PublishMsg(ctx, msg)
	if err != nil {
		m.logger.WithContext(ctx).ErrorW("NATS.JSPublish", "publish failed", "subject", subject, "err", err)
		return nil, err
	}
	m.logger.WithContext(ctx).DebugW(
		"NATS.JSPublish", "message published",
		"subject", subject,
		"data", string(data),
		"headers", msg.Header,
		"pubAck", ack,
	)
	return ack, nil
}

// JSPublishIdempotent 向 JetStream 同步幂等发布消息，通过 msgID 在 Duplicates 窗口内去重。
// 相同 msgID 在去重窗口内重复发布时，服务端只保留第一条，PubAck.Duplicate 为 true。
// 适合需要 exactly-once 语义的场景，如订单创建、支付回调等。
//
// header 可选：不传则使用空 Header，传入则在其基础上追加 Nats-Log-Id。
func (m *Manager) JSPublishIdempotent(ctx context.Context, subject string, data []byte, msgID string, headers ...nats.Header) (*jetstream.PubAck, error) {
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
	ack, err := m.js.PublishMsg(ctx, msg, jetstream.WithMsgID(msgID))
	if err != nil {
		m.logger.WithContext(ctx).ErrorW("NATS.JSPublish", "idempotent publish failed", "subject", subject, "msgId", msgID, "err", err)
		return nil, err
	}

	m.logger.WithContext(ctx).DebugW(
		"NATS.JSPublish", "message published",
		"subject", subject,
		"data", string(data),
		"msgId", msgID,
		"headers", msg.Header,
		"pubAck", ack,
	)
	return ack, nil
}

// JSPublishAsync 向 JetStream 异步发布消息，立即返回 PubAckFuture，不等待服务端 ACK。
// 适合高吞吐场景：批量发布时无需逐条等待确认，吞吐量远高于 JSPublish。
//
// 注意事项：
//   - 持久化失败（ACK 超时等）通过 Manager 初始化时注册的 ErrHandler 回调通知，不在此返回 error
//   - 返回 error 仅表示消息未能入队（pending 超过 MaxPending 上限或连接已关闭），需业务层重试
//   - 需要批量并发等待所有 ACK 时，推荐使用 JSPublishAsyncBatch
//   - 应用退出前 Manager.Close() 会自动等待所有 pending ACK 处理完毕
//
// header 可选：不传则使用空 Header，传入则在其基础上追加 Nats-Log-Id。
func (m *Manager) JSPublishAsync(ctx context.Context, subject string, data []byte, headers ...nats.Header) (jetstream.PubAckFuture, error) {
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
	future, err := m.js.PublishMsgAsync(msg)
	if err != nil {
		m.logger.WithContext(ctx).ErrorW("NATS.JSPublishAsync", "async publish failed", "subject", subject, "err", err)
		return nil, err
	}
	m.logger.WithContext(ctx).DebugW("NATS.JSPublishAsync", "message enqueued", "subject", subject, "data", string(data), "headers", msg.Header)
	return future, nil
}

// JSPublishIdempotentAsync 向 JetStream 异步幂等发布消息，通过 msgID 在 Duplicates 窗口内去重。
// 兼具异步高吞吐和幂等去重能力，适合高频且需要 exactly-once 语义的场景。
// 返回 error 仅表示消息未能入队，ACK 失败通过 ErrHandler 回调通知。
// 需要批量并发等待所有 ACK 时，推荐使用 JSPublishIdempotentAsyncBatch。
//
// header 可选：不传则使用空 Header，传入则在其基础上追加 Nats-Log-Id。
func (m *Manager) JSPublishIdempotentAsync(ctx context.Context, subject string, data []byte, msgID string, headers ...nats.Header) (jetstream.PubAckFuture, error) {
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
	future, err := m.js.PublishMsgAsync(msg, jetstream.WithMsgID(msgID))
	if err != nil {
		m.logger.WithContext(ctx).ErrorW("NATS.JSPublishAsync", "idempotent async publish failed", "subject", subject, "msgId", msgID, "err", err)
		return nil, err
	}
	m.logger.WithContext(ctx).DebugW("NATS.JSPublishAsync", "message enqueued", "subject", subject, "data", string(data), "msgId", msgID, "headers", msg.Header)
	return future, nil
}

// AsyncBatchMsg 是 JSPublishAsyncBatch 的单条消息入参。
// Header 可选，为 nil 时使用空 Header。
type AsyncBatchMsg struct {
	Subject string
	Data    []byte
	Header  nats.Header
}

// JSPublishAsyncBatch 串行批量异步发布，每条消息可指定不同 Subject 和 Header，并发等待所有消息的 ACK。
// 任意一条 ACK 失败则返回第一个错误，其余消息仍继续等待各自的 ACK。
// 适合需要批量发布且关注整体成功与否的场景。
func (m *Manager) JSPublishAsyncBatch(ctx context.Context, msgs []AsyncBatchMsg) error {
	futures := make([]jetstream.PubAckFuture, 0, len(msgs))
	for _, msg := range msgs {
		future, err := m.JSPublishAsync(ctx, msg.Subject, msg.Data, msg.Header)
		if err != nil {
			return err
		}
		futures = append(futures, future)
	}
	g, _ := errgroup.WithContext(ctx)
	for i, f := range futures {
		g.Go(func() error {
			select {
			case <-f.Ok():
				return nil
			case err := <-f.Err():
				m.logger.WithContext(ctx).ErrorW("NATS.JSPublishAsyncBatch", "ack failed", "subject", msgs[i].Subject, "index", i, "err", err)
				return err
			}
		})
	}
	return g.Wait()
}

// IdempotentBatchMsg 是 JSPublishIdempotentAsyncBatch 的单条消息入参，包含目标 Subject、消息体和幂等 ID。
// Header 可选，为 nil 时使用空 Header。
type IdempotentBatchMsg struct {
	Subject string
	Data    []byte
	MsgID   string
	Header  nats.Header
}

// JSPublishIdempotentAsyncBatch 串行批量幂等异步发布，每条消息可指定不同 Subject 和 Header，并发等待所有消息的 ACK。
// 每条消息通过 MsgID 在 Duplicates 窗口内去重，任意一条 ACK 失败则返回第一个错误。
// 适合高频且需要 exactly-once 语义的批量发布场景。
func (m *Manager) JSPublishIdempotentAsyncBatch(ctx context.Context, msgs []IdempotentBatchMsg) error {
	futures := make([]jetstream.PubAckFuture, 0, len(msgs))
	for _, msg := range msgs {
		future, err := m.JSPublishIdempotentAsync(ctx, msg.Subject, msg.Data, msg.MsgID, msg.Header)
		if err != nil {
			return err
		}
		futures = append(futures, future)
	}
	g, _ := errgroup.WithContext(ctx)
	for i, f := range futures {
		g.Go(func() error {
			select {
			case <-f.Ok():
				return nil
			case err := <-f.Err():
				m.logger.WithContext(ctx).ErrorW("NATS.JSPublishIdempotentAsyncBatch", "ack failed", "subject", msgs[i].Subject, "index", i, "err", err)
				return err
			}
		})
	}
	return g.Wait()
}
