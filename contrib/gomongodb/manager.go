package gomongodb

import (
	"context"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

type Manager struct {
	ctx            context.Context
	clientDatabase *mongo.Database
	logger         *log.Helper
	config         *Config
}

func NewManager(ctx context.Context, configPath string, logger *log.Helper) *Manager {
	manager := &Manager{
		ctx:            ctx,
		logger:         logger,
		config:         loadConfig(configPath),
		clientDatabase: nil,
	}
	manager.connect()
	return manager
}

func (m *Manager) ClientDatabase() *mongo.Database {
	return m.clientDatabase
}

func (m *Manager) connect() {
	clientOptions := options.Client().
		ApplyURI("mongodb://" + m.config.Host + ":" + strconv.Itoa(m.config.Port)).
		SetMaxPoolSize(m.config.MaxPoolSize).
		SetMinPoolSize(m.config.MinPoolSize).
		SetMaxConnIdleTime(time.Duration(m.config.MaxConnIdleTimeSec) * time.Second).
		SetMonitor(&event.CommandMonitor{
			Started:   startedMonitorFunc(m.logger),
			Succeeded: succeededMonitorFunc(m.logger),
			Failed:    failedMonitorFunc(m.logger),
		})

	client, err := mongo.Connect(m.ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	m.clientDatabase = client.Database(m.config.Database)
}
