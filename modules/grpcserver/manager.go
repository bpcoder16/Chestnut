package grpcserver

import (
	"context"
	"net"

	"github.com/bpcoder16/Chestnut/v2/logit"
	"google.golang.org/grpc"
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
	manager := &Manager{
		config:     config,
		server:     grpc.NewServer(),
		serverList: serviceList,
	}
	return manager
}

func (m *Manager) Run(ctx context.Context) error {
	if len(m.serverList) > 0 {
		for _, service := range m.serverList {
			service.RegisterService(m.server)
		}
	}

	// 创建 gRPC 监听器
	listen, err := net.Listen("tcp", ":"+m.config.Port)
	if err != nil {
		logit.Context(ctx).FatalW("grpcServer failed to listen: ", err)
		return err
	}

	go func() {
		select {
		case <-ctx.Done():
			logit.Context(ctx).InfoW("grpcServer.Manager.Run", "Context cancelled, preparing to shutdown")
		}

		m.server.GracefulStop()
		logit.Context(ctx).InfoW("grpcServer.Manager.Run", "shutdown completed, exited")
	}()

	logit.Context(ctx).InfoW("grpcServer.Manager.Run", "grpcServer started")
	return m.server.Serve(listen)
}
