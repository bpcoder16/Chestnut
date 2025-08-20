package httpserver

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/bpcoder16/Chestnut/v2/logit"
)

type Router interface {
	RegisterHandler(*Manager)
}

type Manager struct {
	config  *Config
	handler http.Handler
	server  *http.Server
}

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
			ConnState: func() func(conn net.Conn, state http.ConnState) {
				if config.IsOpenConnStateTraceLog {
					return connStateHandler
				}
				return nil
			}(),
			//BaseContext: func(listener net.Listener) context.Context {
			//	ctx := context.Background()
			//	ctx = context.WithValue(ctx, log.DefaultMessageKey, "HTTP")
			//	ctx = context.WithValue(ctx, log.DefaultLogIdKey, utils.UniqueID())
			//	return ctx
			//},
		},
	}
	return manager
}

func (m *Manager) Run(ctx context.Context) error {
	go func() {
		select {
		case <-ctx.Done():
			logit.Context(ctx).InfoW("httpServer.Manager.Run", "Context cancelled, preparing to shutdown")
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.server.Shutdown(shutdownCtx); err != nil {
			logit.Context(ctx).ErrorW("httpServer.Manager.Run", "shutdown failed: "+err.Error())
		}
		logit.Context(ctx).InfoW("httpServer.Manager.Run", "shutdown completed, exited")
	}()

	logit.Context(ctx).InfoW("httpServer.Manager.Run", "HttpServer started")
	return m.server.ListenAndServe()
}

func connStateHandler(conn net.Conn, state http.ConnState) {
	ctx := context.Background()
	switch state {
	case http.StateNew:
		logit.Context(ctx).DebugW("state", "StateNew", "LocalAddr", conn.LocalAddr(), "RemoteAddr", conn.RemoteAddr())
	case http.StateActive:
		logit.Context(ctx).DebugW("state", "StateActive", "LocalAddr", conn.LocalAddr(), "RemoteAddr", conn.RemoteAddr())
	case http.StateIdle:
		logit.Context(ctx).DebugW("state", "StateIdle", "LocalAddr", conn.LocalAddr(), "RemoteAddr", conn.RemoteAddr())
	case http.StateHijacked:
		logit.Context(ctx).DebugW("state", "StateHijacked", "LocalAddr", conn.LocalAddr(), "RemoteAddr", conn.RemoteAddr())
	case http.StateClosed:
		logit.Context(ctx).DebugW("state", "StateClosed", "LocalAddr", conn.LocalAddr(), "RemoteAddr", conn.RemoteAddr())
	}
}
