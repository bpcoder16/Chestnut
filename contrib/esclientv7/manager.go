package esclientv7

import (
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/elastic/go-elasticsearch/v7"
)

type Manager struct {
	logger *log.Helper
	config *Config
	client *elasticsearch.Client
}

func NewManager(configPath string, logger *log.Helper) *Manager {
	manager := &Manager{
		logger: logger,
		config: loadConfig(configPath),
		client: nil,
	}
	manager.connect()
	return manager
}

func (m *Manager) Client() *elasticsearch.Client {
	return m.client
}

func (m *Manager) connect() {
	cfg := elasticsearch.Config{
		Addresses: m.config.Addresses,
		Username:  m.config.Username,
		Password:  m.config.Password,
		// 后续优化可以考虑设置
		//Transport: &http.Transport{},
	}

	if m.config.EnableDebug {
		cfg.Logger = &Logger{
			Helper:      m.logger,
			EnableDebug: m.config.EnableDebug,
		}
	}
	var err error
	m.client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		panic("failed to connect elasticsearch: " + err.Error())
	}
}
