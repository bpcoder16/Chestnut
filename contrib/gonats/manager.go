package gonats

import (
	"crypto/tls"
	"strings"
	"sync"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Manager 管理 NATS 连接、JetStream 上下文和 Stream 注册表。
// 通过 NewManager 创建，生命周期与应用一致。
type Manager struct {
	nc          *nats.Conn
	js          jetstream.JetStream
	jsBaseOpts  []jetstream.JetStreamOpt // connect() 时确定的默认选项，不可变
	streams       sync.Map // key: stream name (string), value: jetstream.Stream
	pullConsumers sync.Map // key: consumer name (string), value: *PullConsumer
	logger      *log.Helper
	config      *Config
}

// NewManager 根据配置文件路径创建 Manager 并建立连接。
// 连接失败时 panic，符合 Chestnut 框架的初始化约定。
func NewManager(configPath string, logger *log.Helper) *Manager {
	m := &Manager{
		logger: logger,
		config: loadConfig(configPath),
	}
	m.connect()
	return m
}

// NC 返回底层 *nats.Conn，用于 Core NATS 操作或高级场景。
func (m *Manager) NC() *nats.Conn {
	return m.nc
}

// JS 返回默认 JetStream 上下文，用于流管理和发布操作。
func (m *Manager) JS() jetstream.JetStream {
	return m.js
}

// ResetJetStream 在默认选项基础上叠加新选项并重建 JetStream 上下文。
// 每次调用均以 connect() 时的默认选项为基准，新选项追加其后覆盖同类默认值，
// 多次调用之间互不累积，不会产生重复条目。
// 仅应在应用初始化阶段、任何发布操作开始之前调用。
// 若已有 PublishAsync 调用在途，会先执行 CleanupPublisher 清理旧上下文的订阅和 pending acks，
// 清理过程会对所有未完成的异步发布触发 ErrHandler。
func (m *Manager) ResetJetStream(opts ...jetstream.JetStreamOpt) error {
	m.js.CleanupPublisher()
	js, err := jetstream.New(m.nc, append(m.jsBaseOpts, opts...)...)
	if err != nil {
		return err
	}
	m.js = js
	return nil
}

// NewJetStreamWithDomain 从当前连接创建指向指定 Domain 的 JetStream 上下文。
// 适用于 NATS 多租户部署，需要跨 Domain 操作 Stream 的场景。
// 返回的上下文应在启动时创建一次后全局复用，不应在请求路径中反复调用此方法。
func (m *Manager) NewJetStreamWithDomain(domain string, opts ...jetstream.JetStreamOpt) (jetstream.JetStream, error) {
	return jetstream.NewWithDomain(m.nc, domain, opts...)
}

// Close 优雅关闭 NATS 连接，等待 in-flight 消息处理完成后再断开。
// 超时时间由配置的 DrainTimeoutMillisecond 决定，超时后强制关闭。
// 应在应用退出时调用（如注册到 cdefer）。
func (m *Manager) Close() {
	if m.nc == nil || m.nc.IsClosed() {
		return
	}
	// 先等待所有 JetStream 异步发布的 pending ACK 处理完毕（超时由 WithPublishAsyncTimeout 控制）。
	// 必须在 Drain 之前调用，否则连接关闭后 pending ACK 直接丢弃。
	<-m.js.PublishAsyncComplete()
	// Drain 超时时间已在 connect 时通过 nats.DrainTimeout() 选项配置，
	// 此处调用会等待订阅侧 in-flight 消息处理完成或超时后返回。
	if err := m.nc.Drain(); err != nil {
		m.logger.WarnW("NATS.Close", "drain failed, connection closed", "err", err)
	}
}

func (m *Manager) connect() {
	opts := []nats.Option{
		nats.Name(m.config.Name),
		nats.MaxReconnects(m.config.MaxReconnects),
		nats.ReconnectBufSize(m.config.ReconnectBufSizeMegabyte * 1024 * 1024),
		nats.ReconnectWait(time.Duration(m.config.ReconnectWaitMillisecond) * time.Millisecond),
		nats.ReconnectJitter(
			time.Duration(m.config.ReconnectJitterMillisecond)*time.Millisecond,
			time.Duration(m.config.ReconnectJitterMillisecond)*time.Millisecond,
		),
		nats.Timeout(time.Duration(m.config.ConnectTimeoutMillisecond) * time.Millisecond),
		nats.DrainTimeout(time.Duration(m.config.DrainTimeoutMillisecond) * time.Millisecond),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				m.logger.ErrorW("NATS.Disconnect", "connection disconnected", "err", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			m.logger.InfoW("NATS.Reconnect", "reconnected successfully", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			m.logger.InfoW("NATS.Closed", "connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			m.logger.ErrorW("NATS.AsyncError", "async error", "subject", sub.Subject, "err", err)
		}),
	}

	if m.config.Token != "" {
		opts = append(opts, nats.Token(m.config.Token))
	} else if m.config.Username != "" {
		opts = append(opts, nats.UserInfo(m.config.Username, m.config.Password))
	}

	if m.config.TLS.enabled() {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: m.config.TLS.InsecureSkipVerify, //nolint:gosec
		}
		opts = append(opts, nats.Secure(tlsCfg))
		if m.config.TLS.CaFile != "" {
			opts = append(opts, nats.RootCAs(m.config.TLS.CaFile))
		}
		if m.config.TLS.CertFile != "" && m.config.TLS.KeyFile != "" {
			opts = append(opts, nats.ClientCert(m.config.TLS.CertFile, m.config.TLS.KeyFile))
		}
	}

	serverURL := strings.Join(m.config.URLs, ",")
	nc, err := nats.Connect(serverURL, opts...)
	if err != nil {
		panic("failed to connect NATS [" + serverURL + "]: " + err.Error())
	}
	m.nc = nc

	m.jsBaseOpts = []jetstream.JetStreamOpt{
		// 异步发布失败时记录错误日志，避免静默丢弃。
		jetstream.WithPublishAsyncErrHandler(func(_ jetstream.JetStream, msg *nats.Msg, err error) {
			m.logger.ErrorW("NATS.AsyncPublish", "async publish failed", "subject", msg.Subject, "err", err)
		}),
		// 异步发布等待 ACK 的超时，超时后触发 ErrHandler。
		// 必须小于 drainTimeoutMillisecond，确保 nc.Drain() 关闭前 pending publish 能触发 ErrHandler。
		jetstream.WithPublishAsyncTimeout(5 * time.Second),
		// 最大 in-flight 异步发布数，超出后 PublishAsync 阻塞。
		jetstream.WithPublishAsyncMaxPending(4000),
		// JetStream 管理类 API 请求超时（ctx 无 deadline 时生效）。
		jetstream.WithDefaultTimeout(5 * time.Second),
	}
	js, err := jetstream.New(nc, m.jsBaseOpts...)
	if err != nil {
		panic("failed to create JetStream context: " + err.Error())
	}
	m.js = js

	m.logger.InfoW("NATS.Connected", "connected successfully", "url", nc.ConnectedUrl())
}
