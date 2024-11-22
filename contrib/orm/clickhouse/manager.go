package clickhouse

import (
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/core/utils"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strconv"
	"time"
)

type Manager struct {
	dbMaster *gorm.DB
	dbSlaves []*gorm.DB
	logger   *log.Helper
	config   *Config
}

func NewManager(configPath string, logger *log.Helper) *Manager {
	manager := &Manager{
		logger:   logger,
		config:   loadConfig(configPath),
		dbMaster: nil,
		dbSlaves: make([]*gorm.DB, 0, 10),
	}
	manager.connectMaster()
	manager.connectSlaves()
	return manager
}

func (m *Manager) MasterDB() *gorm.DB {
	return m.dbMaster
}

func (m *Manager) SlaveDB() *gorm.DB {
	switch len(m.dbSlaves) {
	case 0:
		return m.MasterDB()
	case 1:
		return m.dbSlaves[0]
	default:
		return m.dbSlaves[utils.RandIntN(len(m.dbSlaves))]
	}
}

func (m *Manager) connect(config *ConfigItem) *gorm.DB {
	dsn := "clickhouse://" + config.Username + ":" + config.Password + "@" +
		config.Host + ":" + strconv.Itoa(config.Port) + "/" + config.Database + "?dial_timeout=10s&read_timeout=20s"
	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{
		Logger: NewLogger(m.logger, logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Info,            // Log level
			IgnoreRecordNotFoundError: true,                   // Ignore ErrRecordNotFound error for zaplogger
			ParameterizedQueries:      false,                  // Don't include params in the SQL log
			Colorful:                  false,
		}),
	})
	if err != nil {
		panic(dsn + ", failed to connect database: " + err.Error())
	}
	return db
}

func (m *Manager) setConnectionPool(db *gorm.DB, config *ConfigItem) {
	sqlDB, _ := db.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)
}

func (m *Manager) connectMaster() {
	m.dbMaster = m.connect(m.config.Master)
	m.setConnectionPool(m.dbMaster, m.config.Master)
}

func (m *Manager) connectSlaves() {
	if len(m.config.Slaves) > 0 {
		for _, slaveConfig := range m.config.Slaves {
			db := m.connect(slaveConfig)
			m.setConnectionPool(db, slaveConfig)
			m.dbSlaves = append(m.dbSlaves, db)
		}
	}
}
