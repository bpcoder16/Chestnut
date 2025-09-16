package grpcserver

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/bpcoder16/Chestnut/v2/logit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Service interface {
	RegisterService(s grpc.ServiceRegistrar)
}

type Manager struct {
	config     *Config
	server     *grpc.Server
	serverList []Service
}

func NewManager(configPath string, serviceList ...Service) *Manager {
	config := loadConfig(configPath)

	// 构建服务器选项
	opts := buildServerOptions(config)

	manager := &Manager{
		config:     config,
		server:     grpc.NewServer(opts...),
		serverList: serviceList,
	}
	return manager
}

// buildServerOptions 构建 gRPC 服务器选项
func buildServerOptions(config *Config) []grpc.ServerOption {
	opts := []grpc.ServerOption{
		// 性能配置
		grpc.MaxConcurrentStreams(config.Performance.MaxConcurrentStreams),
		grpc.MaxRecvMsgSize(config.Performance.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.Performance.MaxSendMsgSize),
		grpc.InitialWindowSize(config.Performance.InitialWindowSize),
		grpc.InitialConnWindowSize(config.Performance.InitialConnWindowSize),
		grpc.WriteBufferSize(config.Performance.WriteBufferSize),
		grpc.ReadBufferSize(config.Performance.ReadBufferSize),
		//grpc.NumStreamWorkers(config.Performance.NumStreamWorkers),

		// Keepalive 配置
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     config.Keepalive.MaxConnectionIdleSec * time.Second,
			MaxConnectionAge:      config.Keepalive.MaxConnectionAgeSec * time.Second,
			MaxConnectionAgeGrace: config.Keepalive.MaxConnectionAgeGraceSec * time.Second,
			Time:                  config.Keepalive.TimeSec * time.Second,
			Timeout:               config.Keepalive.TimeoutSec * time.Second,
		}),
		// 连接策略配置
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.KeepAlivePolicy.MinTimeSec * time.Second,
			PermitWithoutStream: config.KeepAlivePolicy.PermitWithoutStream,
		}),
	}

	//// 如果启用了 TLS，添加 TLS 凭证
	//if config.TLS.Enabled {
	//	if config.TLS.CertFile != "" && config.TLS.KeyFile != "" {
	//		creds, err := credentials.NewServerTLSFromFile(config.TLS.CertFile, config.TLS.KeyFile)
	//		if err != nil {
	//			panic("Failed to load TLS credentials: " + err.Error())
	//		}
	//		opts = append(opts, grpc.Creds(creds))
	//	}
	//}

	return opts
}

func (m *Manager) Run(ctx context.Context) error {
	if len(m.serverList) > 0 {
		for _, service := range m.serverList {
			service.RegisterService(m.server)
		}
	}

	// 启动优雅关闭监听器
	go m.gracefulShutdown(ctx)

	// 创建 gRPC 监听器
	listen, err := net.Listen("tcp", ":"+m.config.Port)
	if err != nil {
		logit.Context(ctx).FatalW("grpcServer failed to listen: ", err)
		return err
	}

	logit.Context(ctx).InfoW("grpcServer.Manager.Run", "grpcServer started", "port", m.config.Port)

	// 区分正常关闭和异常错误
	if errS := m.server.Serve(listen); errS != nil && !errors.Is(errS, grpc.ErrServerStopped) {
		return errS
	}
	return nil
}

// gracefulShutdown 处理优雅关闭逻辑
func (m *Manager) gracefulShutdown(ctx context.Context) {
	// 等待context取消信号
	<-ctx.Done()
	logit.Context(ctx).InfoW("grpcServer.Manager.Run", "Context cancelled, preparing to shutdown")

	m.server.GracefulStop()
	logit.Context(ctx).InfoW("grpcServer.Manager.Run", "shutdown completed successfully")
}
