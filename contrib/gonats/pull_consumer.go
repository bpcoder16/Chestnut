package gonats

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// PullConsumer 封装 JetStream Pull Consumer，提供多种消费模式。
// Pull Consumer 由客户端主动拉取消息，适合需要控制消费速率、批处理的场景。
type PullConsumer struct {
	consumer jetstream.Consumer
	logger   *log.Helper
}

// Consumer 返回底层 jetstream.Consumer，用于直接调用原生 API。
func (c *PullConsumer) Consumer() jetstream.Consumer {
	return c.consumer
}

// MustRegisterPullConsumer 创建或更新 Pull Consumer 并保存到全局注册表，供后续通过 GetPullConsumer 取用。
// 失败时直接 panic，用于应用启动阶段初始化。
// 注册 key 优先使用 cfg.Durable，其次 cfg.Name。
func (m *Manager) MustRegisterPullConsumer(ctx context.Context, streamName string, cfg PullConsumerConfig) *PullConsumer {
	ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "NATS")
	consumer, err := m.js.CreateOrUpdateConsumer(ctx, streamName, buildPullConsumerConfig(cfg))
	if err != nil {
		key := cfg.Durable
		if key == "" {
			key = cfg.Name
		}
		panic("failed to register pull consumer [" + key + "] on stream [" + streamName + "]: " + err.Error())
	}
	pc := &PullConsumer{consumer: consumer, logger: m.logger}
	key := cfg.Durable
	if key == "" {
		key = cfg.Name
	}
	m.pullConsumers.Store(key, pc)
	m.logger.WithContext(ctx).InfoW("NATS.MustRegisterPullConsumer", "pull consumer registered",
		"stream", streamName, "consumer", key)
	return pc
}

// GetPullConsumer 从注册表中取出已注册的 Pull Consumer。
// 若 Consumer 尚未通过 RegisterPullConsumer 注册，返回 error。
func (m *Manager) GetPullConsumer(name string) (*PullConsumer, error) {
	v, ok := m.pullConsumers.Load(name)
	if !ok {
		return nil, fmt.Errorf("pull consumer %q not registered, call RegisterPullConsumer first", name)
	}
	return v.(*PullConsumer), nil
}

// Fetch 一次性拉取最多 batch 条消息，阻塞等待直到凑满或 maxWait 超时。
//
// 使用示例（定时批处理任务）：
//
//	msgs, err := consumer.Fetch(ctx, 100, 10*time.Second, 0)
//	if err != nil { ... }
//	for msg := range msgs.Messages() {
//	    process(msg.Data())
//	    msg.Ack()
//	}
//	if msgs.Error() != nil { ... }
func (c *PullConsumer) Fetch(ctx context.Context, batch int, maxWait time.Duration) (jetstream.MessageBatch, error) {
	ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "NATS")
	c.logger.WithContext(ctx).DebugW("NATS.Fetch", "fetching messages",
		"batch", batch, "maxWait", maxWait)
	msgs, err := c.consumer.Fetch(batch,
		jetstream.FetchMaxWait(maxWait),
	)
	if err != nil {
		c.logger.WithContext(ctx).ErrorW("NATS.Fetch", "fetch failed",
			"batch", batch, "maxWait", maxWait, "err", err)
		return nil, err
	}
	return msgs, nil
}

// FetchNoWait 立即返回当前可用消息，最多 batch 条，无消息则返回空，不阻塞等待。
// 适合不希望阻塞的轮询场景。
//
// 使用示例：
//
//	msgs, err := consumer.FetchNoWait(ctx, 50)
//	if err != nil { ... }
//	for msg := range msgs.Messages() {
//	    msg.Ack()
//	}
func (c *PullConsumer) FetchNoWait(ctx context.Context, batch int) (jetstream.MessageBatch, error) {
	ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "NATS")
	c.logger.WithContext(ctx).DebugW("NATS.FetchNoWait", "fetching messages", "batch", batch)
	msgs, err := c.consumer.FetchNoWait(batch)
	if err != nil {
		c.logger.WithContext(ctx).ErrorW("NATS.FetchNoWait", "fetch failed", "batch", batch, "err", err)
		return nil, err
	}
	return msgs, nil
}

// FetchBytes 按字节上限拉取消息，最多拉取 maxBytes 字节，不会截断单条消息。
// 适合消息体大小不均匀、需要控制单批内存占用的场景。
//
// 使用示例：
//
//	msgs, err := consumer.FetchBytes(ctx, 1*1024*1024, 10*time.Second, 0)
//	for msg := range msgs.Messages() {
//	    msg.Ack()
//	}
func (c *PullConsumer) FetchBytes(ctx context.Context, maxBytes int, maxWait time.Duration, minPending int) (jetstream.MessageBatch, error) {
	ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "NATS")
	c.logger.WithContext(ctx).DebugW("NATS.FetchBytes", "fetching messages",
		"maxBytes", maxBytes, "maxWait", maxWait, "minPending", minPending)
	msgs, err := c.consumer.FetchBytes(maxBytes,
		jetstream.FetchMaxWait(maxWait),
	)
	if err != nil {
		c.logger.WithContext(ctx).ErrorW("NATS.FetchBytes", "fetch failed",
			"maxBytes", maxBytes, "maxWait", maxWait, "minPending", minPending, "err", err)
		return nil, err
	}
	return msgs, nil
}

// Consume 以回调方式持续消费消息，后台自动拉取并将消息推给 handler。
// 这是长期运行服务的推荐模式：消息来一条处理一条，无需手动管理拉取节奏。
// 返回的 ConsumeContext 可调用 Stop() 停止消费。
//
// maxMessages 指定客户端预取缓冲的最大消息条数，0 使用服务端默认值（500）。
// errHandler 指定消费异常回调（心跳超时、Consumer 被删除等），nil 时默认记录 Error 日志。
// handler 接收 ctx 和消息，ctx 中已注入发布方传递的 logId，可直接用于日志打印。
//
// 使用示例（在 g.Go 中启动）：
//
//	consCtx, err := consumer.Consume(ctx, func(msgCtx context.Context, msg jetstream.Msg) {
//	    logit.Context(msgCtx).InfoW("MyConsumer", "收到消息", "data", string(msg.Data()))
//	    msg.Ack()
//	}, 200, nil)
//	if err != nil { return err }
//	defer consCtx.Stop()
func (c *PullConsumer) Consume(ctx context.Context, handler func(context.Context, jetstream.Msg), maxMessages int, errHandler jetstream.ConsumeErrHandlerFunc) (jetstream.ConsumeContext, error) {
	ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "NATS")
	opts := make([]jetstream.PullConsumeOpt, 0, 2)
	if maxMessages > 0 {
		opts = append(opts, jetstream.PullMaxMessages(maxMessages))
	}
	if errHandler != nil {
		opts = append(opts, jetstream.ConsumeErrHandler(errHandler))
	} else {
		opts = append(opts, jetstream.ConsumeErrHandler(func(_ jetstream.ConsumeContext, err error) {
			if errors.Is(err, jetstream.ErrConsumerDeleted) || errors.Is(err, jetstream.ErrBadRequest) {
				// 终止性错误：Consumer 已停止，需调用方重建
				c.logger.WithContext(ctx).ErrorW("NATS.Consume", "terminal error, consumer stopped", "err", err)
				return
			}
			// 非终止性错误（心跳超时、Leader 切换等）：库自动恢复，记录 Warn 即可
			c.logger.WithContext(ctx).WarnW("NATS.Consume", "transient error, will recover", "err", err)
		}))
	}
	c.logger.WithContext(ctx).DebugW("NATS.Consume", "starting consumer", "maxMessages", maxMessages)
	consCtx, err := c.consumer.Consume(func(msg jetstream.Msg) {
		msgCtx := extractLogIdFromJSMsg(context.Background(), msg)
		c.logger.WithContext(msgCtx).DebugW("NATS.Consume", "received message",
			"subject", msg.Subject(), "data", string(msg.Data()), "headers", msg.Headers())
		handler(msgCtx, msg)
	}, opts...)
	if err != nil {
		c.logger.WithContext(ctx).ErrorW("NATS.Consume", "failed to start consumer", "err", err)
		return nil, err
	}
	return consCtx, nil
}

// Messages 返回消息迭代器，通过 iter.Next() 逐条消费（阻塞直到有消息）。
// 适合需要手动控制消费节奏或实现自定义并发模型的场景。
// 返回的 MessagesContext 可调用 Stop() 停止迭代。
//
// 可选配置：
//   - jetstream.PullMaxMessages(n)：内存中缓冲的最大消息数
//   - jetstream.PullMaxBytes(n)：内存中缓冲的最大字节数
//
// 使用示例（顺序处理）：
//
//	iter, err := consumer.Messages()
//	defer iter.Stop()
//	for {
//	    msg, err := iter.Next()
//	    if err != nil { break }
//	    process(msg.Data())
//	    msg.Ack()
//	}
func (c *PullConsumer) Messages(opts ...jetstream.PullMessagesOpt) (jetstream.MessagesContext, error) {
	return c.consumer.Messages(opts...)
}

// ConsumeWithWorkers 使用固定数量的 goroutine 并发消费消息。
// 适合消息处理耗时较长、需要并发提升吞吐量，同时又要限制最大并发数的场景。
//
// maxMessages 控制客户端预取缓冲大小，影响 ACK 超时风险与吞吐量的平衡：
//   - 传 0 时默认为 1（安全默认值，适合大多数 worker 场景）
//   - handler 处理远快于 AckWait（如 < AckWait/10）时可适当调大以减少网络往返，提升吞吐量
//   - handler 处理耗时接近 AckWait 时必须设为 1，否则缓冲中等待的消息会超时触发重复投递
//
// 实现原理：
//  1. 创建指定 maxMessages 的消息迭代器
//  2. 用 semaphore（带缓冲 channel）限制同时运行的 worker 数量
//  3. 每个 worker goroutine 拉取并处理一条消息
//  4. ctx 取消时停止迭代器，等待所有 worker 完成后返回
//
// handler 内部应自行调用 msg.Ack() / msg.Nak() / msg.Term()。
//
// 使用示例（在 g.Go 中启动）：
//
//	g.Go(func() error {
//	    return consumer.ConsumeWithWorkers(ctx, 10, 0, func(msg jetstream.Msg) {
//	        if err := process(msg.Data()); err != nil {
//	            msg.NakWithDelay(5 * time.Second)
//	            return
//	        }
//	        msg.Ack()
//	    })
//	})
func (c *PullConsumer) ConsumeWithWorkers(ctx context.Context, numWorkers int, maxMessages int, handler func(msg jetstream.Msg)) error {
	return c.ConsumeWithWorkersContext(ctx, numWorkers, maxMessages, func(_ context.Context, msg jetstream.Msg) {
		handler(msg)
	})
}

// ConsumeWithWorkersContext 使用固定数量的 goroutine 并发消费消息，并为每条消息注入消息级 context。
//
// handler 接收的 msgCtx 中已从消息 Header 还原发布方传递的 logId；若 Header 中没有 logId，则自动生成。
// 调用方应使用 msgCtx 打印消息处理日志，避免多个消息共用启动协程的外层 ctx。
func (c *PullConsumer) ConsumeWithWorkersContext(ctx context.Context, numWorkers int, maxMessages int, handler func(context.Context, jetstream.Msg)) error {
	ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "NATS")
	if numWorkers <= 0 {
		numWorkers = 5
	}
	if maxMessages <= 0 {
		maxMessages = 1
	}
	iter, err := c.consumer.Messages(jetstream.PullMaxMessages(maxMessages))
	if err != nil {
		c.logger.WithContext(ctx).ErrorW("NATS.ConsumeWithWorkers", "failed to create messages iterator", "err", err)
		return err
	}

	c.logger.WithContext(ctx).InfoW("NATS.ConsumeWithWorkers", "subscribed successfully",
		"numWorkers", numWorkers, "maxMessages", maxMessages)

	// ctx 取消时停止迭代器，令 iter.Next() 返回 ErrMsgIteratorClosed
	go func() {
		<-ctx.Done()
		iter.Stop()
	}()

	var (
		wg      sync.WaitGroup
		sem     = make(chan struct{}, numWorkers)
		iterErr error
	)

	for {
		// 获取 semaphore slot，阻塞直到有空余 worker
		sem <- struct{}{}

		msg, nextErr := iter.Next()
		if nextErr != nil {
			// 释放刚获取的 slot，不再启动 worker
			<-sem
			if !errors.Is(nextErr, jetstream.ErrMsgIteratorClosed) {
				// 非正常关闭的错误（如心跳超时）保留，让上层决定是否重启
				c.logger.WithContext(ctx).ErrorW("NATS.ConsumeWithWorkers", "iterator error", "err", nextErr)
				iterErr = nextErr
			}
			break
		}

		wg.Add(1)
		go func(m jetstream.Msg) {
			defer func() {
				wg.Done()
				<-sem
			}()
			msgCtx := buildPullConsumerMessageContext(ctx, m.Headers())
			c.logger.WithContext(msgCtx).InfoW(
				"NATS.Action", "ConsumeWithWorkers.ReceivedMessage",
				"subject", m.Subject(),
				"data", string(m.Data()),
				"headers", m.Headers(),
			)
			handler(msgCtx, m)
		}(msg)
	}

	wg.Wait()
	c.logger.WithContext(ctx).DebugW("NATS.ConsumeWithWorkers", "all workers stopped")
	if iterErr != nil {
		return iterErr
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}

func buildPullConsumerMessageContext(parent context.Context, header nats.Header) context.Context {
	return extractLogIdFromHeader(parent, header)
}
