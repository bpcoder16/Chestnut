package httpserver

import (
	"context"
	"github.com/bpcoder16/Chestnut/logit"
	"net"
	"net/http"
	"time"
)

type Router interface {
	RegisterHandler(*Manager)
}

type Manager struct {
	config  *Config
	handler http.Handler
}

func NewManager(configPath string, handler http.Handler) *Manager {
	manager := &Manager{
		config:  loadConfig(configPath),
		handler: handler,
	}
	return manager
}

func (m *Manager) Run() error {
	return (&http.Server{
		Addr:              ":" + m.config.Port,
		Handler:           m.handler,
		ReadTimeout:       m.config.ReadTimeoutMillisecond * time.Millisecond,
		ReadHeaderTimeout: m.config.ReadHeaderTimeoutMillisecond * time.Millisecond,
		WriteTimeout:      m.config.WriteTimeoutMillisecond * time.Millisecond,
		IdleTimeout:       m.config.IdleTimeoutMillisecond * time.Millisecond,
		MaxHeaderBytes:    m.config.MaxHeaderBytes,
		ConnState: func() func(conn net.Conn, state http.ConnState) {
			if m.config.IsOpenConnStateTraceLog {
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
	}).ListenAndServe()
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
