package grpcserver

import (
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/utils"
)

type Config struct {
	Port string

	// 性能配置
	Performance struct {
		MaxConcurrentStreams  uint32
		MaxRecvMsgSize        int
		MaxSendMsgSize        int
		InitialWindowSize     int32
		InitialConnWindowSize int32
		WriteBufferSize       int
		ReadBufferSize        int
		//NumStreamWorkers uint32
	}

	// Keepalive 配置
	Keepalive struct {
		MaxConnectionIdleSec     time.Duration
		MaxConnectionAgeSec      time.Duration
		MaxConnectionAgeGraceSec time.Duration
		TimeSec                  time.Duration
		TimeoutSec               time.Duration
	}

	// 连接策略配置
	KeepAlivePolicy struct {
		MinTimeSec          time.Duration
		PermitWithoutStream bool
	}

	//TLS struct {
	//	Enabled  bool
	//	CertFile string
	//	KeyFile  string
	//}
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load grpc Server conf err:" + err.Error())
	}
	return &config
}
