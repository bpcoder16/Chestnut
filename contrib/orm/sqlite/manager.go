package sqlite

import (
	"time"

	"github.com/bpcoder16/Chestnut/v2/core/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Manager struct {
	db     *gorm.DB
	logger *log.Helper
	config *Config
}

func NewManager(configPath string, logger *log.Helper) *Manager {
	manager := &Manager{
		db:     nil,
		logger: logger,
		config: loadConfig(configPath),
	}
	manager.db = manager.connect()
	manager.setConnectionPool()
	return manager
}

func (m *Manager) DB() *gorm.DB {
	return m.db
}

func (m *Manager) connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(m.config.DSN), &gorm.Config{
		Logger: NewLogger(m.logger, logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Info,            // Log level
			IgnoreRecordNotFoundError: true,                   // Ignore ErrRecordNotFound error for zaplogger
			ParameterizedQueries:      false,                  // Don't include params in the SQL log
			Colorful:                  false,
		}),
	})

	if err != nil {
		panic(m.config.DSN + ", failed to connect database: " + err.Error())
	}
	return db
}

func (m *Manager) setConnectionPool() {
	sqlDB, _ := m.db.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(m.config.MaxIdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(m.config.MaxOpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(0)
}
