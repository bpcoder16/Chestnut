package httpserver

import (
	"net/http"
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

func (m *Manager) Run() {
	err := (&http.Server{
		Addr:    ":" + m.config.Server.Port,
		Handler: m.handler,
	}).ListenAndServe()

	panic(err)
}
