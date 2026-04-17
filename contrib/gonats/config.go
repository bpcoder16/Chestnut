package gonats

import (
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// TLSConfig 是 NATS TLS 安全连接配置，使用 tls:// URL 时按需填写。
// 不使用 TLS 时整个 tls 块保持默认即可，无需修改任何字段。
type TLSConfig struct {
	// CertFile 客户端证书路径（双向 mTLS 时使用，需与 KeyFile 一起填写）。
	CertFile string `json:"certFile"`
	// KeyFile 客户端私钥路径（与 CertFile 配对）。
	KeyFile string `json:"keyFile"`
	// CaFile CA 根证书路径，用于校验服务端证书。
	// 服务端使用自签证书时，填写对应 CA；使用公信 CA 签发的证书时可留空。
	CaFile string `json:"caFile"`
	// InsecureSkipVerify 跳过服务端证书校验，仅限开发调试使用，生产环境禁止设为 true。
	// 注意：此字段不会触发 TLS 模式，必须同时配置 CaFile 或 CertFile 才会生效。
	InsecureSkipVerify bool `json:"insecureSkipVerify"`
}

// enabled 判断是否需要附加 TLS 配置到连接选项。
// 仅当提供了证书文件时才触发，InsecureSkipVerify 单独设为 true 不会开启 TLS 模式，
// 避免用户误配置时意外将 nats:// 连接强制升级为 TLS 导致连接失败。
func (t TLSConfig) enabled() bool {
	return t.CertFile != "" || t.KeyFile != "" || t.CaFile != ""
}

// Config 是 NATS 连接配置，对应 nats.yaml。
type Config struct {
	// URLs 是 NATS server 地址列表，支持集群模式。
	// 若为空，默认使用 nats://127.0.0.1:4222。
	// TLS 连接使用 tls:// 前缀。
	URLs []string `json:"urls"`

	// Name 是客户端名称，便于在 NATS 监控面板中识别连接来源。
	Name string `json:"name"`

	// Username / Password / Token 三选一用于认证，均为空则不认证。
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`

	// MaxReconnects 最大重连次数，-1 表示无限重连。默认 10。
	MaxReconnects int `json:"maxReconnects"`

	// ReconnectWaitMillisecond 每次重连等待时间（毫秒整数）。默认 2000。
	ReconnectWaitMillisecond int `json:"reconnectWaitMillisecond"`

	// ReconnectJitterMillisecond 重连等待的随机抖动上限（毫秒整数）。
	// 多实例同时断线时，抖动可避免所有实例同一时刻涌入重连请求。默认 500。
	ReconnectJitterMillisecond int `json:"reconnectJitterMillisecond"`

	// ConnectTimeoutMillisecond 初次连接超时时间（毫秒整数）。默认 5000。
	ConnectTimeoutMillisecond int `json:"connectTimeoutMillisecond"`

	// DrainTimeoutMillisecond 优雅关闭时等待 in-flight 消息处理完成的超时时间（毫秒整数）。
	// 超时后强制关闭，未 ACK 的消息由 server 重投递。默认 5000。
	DrainTimeoutMillisecond int `json:"drainTimeoutMillisecond"`

	// ReconnectBufSizeMegabyte 重连期间缓冲出站消息的内存上限（MB 整数）。
	// 超出后 Publish/JSPublish/PublishAsync 立即返回 ErrReconnectBufExceeded，消息丢弃。
	// 建议根据「单次重连最长时间 × 业务发布吞吐」估算，默认 32。
	ReconnectBufSizeMegabyte int `json:"reconnectBufSizeMegabyte"`

	// TLS 安全连接配置，使用 tls:// URL 时填写。
	TLS TLSConfig `json:"tls"`
}

// StreamConfig 是 JetStream Stream 的创建配置，包含合理的默认值。
type StreamConfig struct {
	// Name 是 Stream 名称（必填），全大写，如 "ORDERS"。
	Name string

	// Subjects 是该 Stream 捕获的 Subject 列表，如 ["ORDERS.*"]。
	Subjects []string

	// Description 可选描述信息。
	Description string

	// Retention 消息保留策略。
	//   LimitsPolicy（默认）：按容量/时间/条数限制保留
	//   InterestPolicy：有 consumer 才保留
	//   WorkQueuePolicy：消息被消费后即删除
	Retention jetstream.RetentionPolicy

	// Storage 存储类型。
	//   FileStorage（默认）：落盘持久化
	//   MemoryStorage：纯内存，重启丢失
	Storage jetstream.StorageType

	// Replicas 副本数，生产环境建议 3。默认 1。
	Replicas int

	// MaxAge 消息最长保留时间，0 表示不限制。建议根据业务设置，如 24h。
	MaxAge time.Duration

	// MaxMsgs 最大消息条数，0 表示不限制。
	MaxMsgs int64

	// MaxBytes 最大字节数，0 表示不限制。
	MaxBytes int64

	// MaxMsgSize 单条消息最大字节数，0 表示不限制。
	MaxMsgSize int32

	// Discard 超出限制时的丢弃策略。
	//   DiscardOld（默认）：丢弃最旧的消息
	//   DiscardNew：拒绝新消息（返回错误）
	Discard jetstream.DiscardPolicy

	// MaxMsgsPerSubject 每个 Subject 最多保留的消息条数，0 表示不限制。
	// 适合"最新状态缓存"场景：设为 1 时每个 Subject 只保留最新一条，类似 key-value store。
	MaxMsgsPerSubject int64

	// Duplicates 消息去重窗口时长，窗口内相同 Msg-Id 只保留一条。默认 2min。
	Duplicates time.Duration
}

// PullConsumerConfig 是 JetStream Pull Consumer 的创建配置。
// Pull Consumer 由客户端主动拉取消息，支持批量、流控、并发等模式。
type PullConsumerConfig struct {
	// Durable 持久消费者名称，设置后重启服务消费进度不丢失（推荐设置）。
	// 与 Name 互斥，Durable 优先。
	Durable string

	// Name 临时消费者名称（Ephemeral），连接断开后自动清理。
	// 若 Durable 已设置则此字段忽略。
	Name string

	// Description 可选描述。
	Description string

	// DeliverPolicy 决定从哪里开始消费。
	//   DeliverAllPolicy（默认）：从 Stream 最早的消息开始
	//   DeliverLastPolicy：只取最新一条
	//   DeliverNewPolicy：只消费订阅后的新消息
	//   DeliverByStartSequencePolicy：从指定序列号开始（需配合 OptStartSeq）
	//   DeliverByStartTimePolicy：从指定时间开始（需配合 OptStartTime）
	DeliverPolicy jetstream.DeliverPolicy

	// OptStartSeq 配合 DeliverByStartSequencePolicy 使用。
	OptStartSeq uint64

	// OptStartTime 配合 DeliverByStartTimePolicy 使用。
	OptStartTime *time.Time

	// FilterSubjects 多 Subject 过滤器（与 FilterSubject 互斥，服务端 2.10+ 支持）。
	FilterSubjects []string

	// AckPolicy 消息确认策略。
	//   AckExplicitPolicy（默认）：每条消息需显式调用 msg.Ack()
	//   AckAllPolicy：确认某条时，其前所有消息一并确认
	//   AckNonePolicy：无需确认，消息投递即视为成功
	AckPolicy jetstream.AckPolicy

	// AckWait 等待 ACK 的超时时间，超时后消息重新投递。默认 30s。
	AckWait time.Duration

	// MaxDeliver 最大投递次数，超过后消息进入 Dead Letter（或丢弃）。
	// -1 表示无限重试，默认 3。
	MaxDeliver int

	// BackOff 自定义重投递退避策略，如 []time.Duration{5s, 30s, 5min}。
	// 若设置，MaxDeliver 自动调整为 len(BackOff)+1。
	BackOff []time.Duration

	// MaxAckPending 最大待确认消息数，控制服务端推送速率。默认 1000。
	// -1 表示无限制。
	MaxAckPending int

	// InactiveThreshold Consumer 空闲超过此时间后服务端自动清理。
	// Durable Consumer 默认不清理，多实例部署时建议设置（如 5min），防止僵尸 Consumer 堆积。
	// 0 表示继承 Stream 或账户配置。
	InactiveThreshold time.Duration

	// ReplayPolicy 消息投递速率策略。
	//   ReplayInstantPolicy（默认）：尽快投递
	//   ReplayOriginalPolicy：按消息写入 Stream 时的原始速率投递，适合回放生产流量
	ReplayPolicy jetstream.ReplayPolicy

	// MaxRequestBatch 服务端限制单次 Pull 请求最多拉取的消息条数。
	// 0 表示不限制，与 MaxRequestMaxBytes 同时设置时取先到者。
	MaxRequestBatch int

	// MaxRequestExpires 服务端限制单次 Pull 请求的最长等待时间。
	// 0 表示不限制。
	MaxRequestExpires time.Duration

	// MaxRequestMaxBytes 服务端限制单次 Pull 请求最多拉取的字节数。
	// 0 表示不限制，与 MaxRequestBatch 同时设置时取先到者。
	MaxRequestMaxBytes int
}

func loadConfig(configPath string) *Config {
	var config Config
	if err := utils.ParseFile(configPath, &config); err != nil {
		panic("load NATS conf err: " + err.Error())
	}
	if len(config.URLs) == 0 {
		config.URLs = []string{nats.DefaultURL}
	}
	if config.MaxReconnects == 0 {
		config.MaxReconnects = 10
	}
	if config.ReconnectWaitMillisecond == 0 {
		config.ReconnectWaitMillisecond = 2000
	}
	if config.ReconnectJitterMillisecond == 0 {
		config.ReconnectJitterMillisecond = 500
	}
	if config.ConnectTimeoutMillisecond == 0 {
		config.ConnectTimeoutMillisecond = 5000
	}
	if config.DrainTimeoutMillisecond == 0 {
		config.DrainTimeoutMillisecond = 5000
	}
	if config.ReconnectBufSizeMegabyte == 0 {
		config.ReconnectBufSizeMegabyte = 32
	}
	return &config
}

func buildStreamConfig(cfg StreamConfig) jetstream.StreamConfig {
	if cfg.Replicas == 0 {
		cfg.Replicas = 1
	}
	if cfg.Duplicates == 0 {
		cfg.Duplicates = 2 * time.Minute
	}
	if cfg.MaxAge == 0 && cfg.MaxMsgs == 0 && cfg.MaxBytes == 0 && cfg.MaxMsgsPerSubject == 0 {
		panic("StreamConfig [" + cfg.Name + "] must set at least one of MaxAge / MaxMsgs / MaxBytes / MaxMsgsPerSubject to prevent unbounded growth")
	}
	return jetstream.StreamConfig{
		Name:              cfg.Name,
		Subjects:          cfg.Subjects,
		Description:       cfg.Description,
		Retention:         cfg.Retention,
		Storage:           cfg.Storage,
		Replicas:          cfg.Replicas,
		MaxAge:            cfg.MaxAge,
		MaxMsgs:           cfg.MaxMsgs,
		MaxBytes:          cfg.MaxBytes,
		MaxMsgSize:        cfg.MaxMsgSize,
		MaxMsgsPerSubject: cfg.MaxMsgsPerSubject,
		Discard:           cfg.Discard,
		Duplicates:        cfg.Duplicates,
	}
}

func buildPullConsumerConfig(cfg PullConsumerConfig) jetstream.ConsumerConfig {
	if cfg.AckWait == 0 {
		cfg.AckWait = 30 * time.Second
	}
	if cfg.MaxDeliver == 0 {
		cfg.MaxDeliver = 3
	}
	if cfg.MaxAckPending == 0 {
		cfg.MaxAckPending = 1000
	}
	jsCfg := jetstream.ConsumerConfig{
		Durable:            cfg.Durable,
		Name:               cfg.Name,
		Description:        cfg.Description,
		DeliverPolicy:      cfg.DeliverPolicy,
		OptStartSeq:        cfg.OptStartSeq,
		OptStartTime:       cfg.OptStartTime,
		FilterSubjects:     cfg.FilterSubjects,
		AckPolicy:          cfg.AckPolicy,
		AckWait:            cfg.AckWait,
		MaxDeliver:         cfg.MaxDeliver,
		BackOff:            cfg.BackOff,
		MaxAckPending:      cfg.MaxAckPending,
		InactiveThreshold:  cfg.InactiveThreshold,
		ReplayPolicy:       cfg.ReplayPolicy,
		MaxRequestBatch:    cfg.MaxRequestBatch,
		MaxRequestExpires:  cfg.MaxRequestExpires,
		MaxRequestMaxBytes: cfg.MaxRequestMaxBytes,
	}
	return jsCfg
}
