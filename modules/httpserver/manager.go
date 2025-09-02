package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/bpcoder16/Chestnut/v2/logit"
)

type Router interface {
	RegisterHandler(*Manager)
}

// Manager HTTP服务器管理器，负责HTTP服务器的创建、配置和生命周期管理
type Manager struct {
	config  *Config
	handler http.Handler
	server  *http.Server
	ctx     context.Context // 添加context字段用于connStateHandler
}

// NewManager 创建新的HTTP服务器管理器
// configPath: 配置文件路径
// handler: HTTP处理器
func NewManager(configPath string, handler http.Handler) *Manager {
	config := loadConfig(configPath)
	manager := &Manager{
		config:  config,
		handler: handler,
		server: &http.Server{
			Addr:              ":" + config.Port,
			Handler:           handler,
			ReadTimeout:       config.ReadTimeoutMillisecond * time.Millisecond,
			ReadHeaderTimeout: config.ReadHeaderTimeoutMillisecond * time.Millisecond,
			WriteTimeout:      config.WriteTimeoutMillisecond * time.Millisecond,
			IdleTimeout:       config.IdleTimeoutMillisecond * time.Millisecond,
			MaxHeaderBytes:    config.MaxHeaderBytes,
		},
	}

	// 设置ConnState处理器（如果启用了连接状态日志）
	if config.IsOpenConnStateTraceLog {
		manager.server.ConnState = manager.connStateHandler
	}

	return manager
}

// Run 启动HTTP服务器并阻塞直到服务器关闭或发生错误
// ctx: 用于控制服务器生命周期的context
func (m *Manager) Run(ctx context.Context) error {
	m.ctx = ctx // 保存context供其他方法使用

	// 启动优雅关闭监听器
	go m.gracefulShutdown(ctx)

	logit.Context(ctx).InfoW("httpServer.Manager.Run", "HttpServer started", "port", m.config.Port)

	// 区分正常关闭和异常错误
	if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// gracefulShutdown 处理优雅关闭逻辑
func (m *Manager) gracefulShutdown(ctx context.Context) {
	// 等待context取消信号
	<-ctx.Done()
	logit.Context(ctx).InfoW("httpServer.Manager.Run", "Context cancelled, preparing to shutdown")

	// 使用配置的关闭超时时间，默认5秒
	shutdownTimeout := 5 * time.Second
	if m.config.ShutdownTimeoutSecond > 0 {
		shutdownTimeout = time.Duration(m.config.ShutdownTimeoutSecond) * time.Second
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := m.server.Shutdown(shutdownCtx); err != nil {
		logit.Context(ctx).ErrorW("httpServer.Manager.Run", "shutdown failed", "error", err.Error())
		return
	}
	logit.Context(ctx).InfoW("httpServer.Manager.Run", "shutdown completed successfully")
}

// connStateHandler 连接状态处理器，记录连接状态变化
func (m *Manager) connStateHandler(conn net.Conn, state http.ConnState) {
	// 使用保存的context而不是Background
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	switch state {
	case http.StateNew:
		logit.Context(ctx).DebugW("httpServer.connState", "StateNew", "local", conn.LocalAddr().String(), "remote", conn.RemoteAddr().String())
	case http.StateActive:
		logit.Context(ctx).DebugW("httpServer.connState", "StateActive", "local", conn.LocalAddr().String(), "remote", conn.RemoteAddr().String())
	case http.StateIdle:
		logit.Context(ctx).DebugW("httpServer.connState", "StateIdle", "local", conn.LocalAddr().String(), "remote", conn.RemoteAddr().String())
	case http.StateHijacked:
		logit.Context(ctx).DebugW("httpServer.connState", "StateHijacked", "local", conn.LocalAddr().String(), "remote", conn.RemoteAddr().String())
	case http.StateClosed:
		logit.Context(ctx).DebugW("httpServer.connState", "StateClosed", "local", conn.LocalAddr().String(), "remote", conn.RemoteAddr().String())
	}
}
