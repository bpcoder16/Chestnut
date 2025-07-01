package mysql

import (
	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/url"
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
	params := url.Values{}
	params.Set("charset", "utf8")
	params.Set("parseTime", "true")
	params.Set("loc", env.TimeLocation().String())
	dsn := config.Username + ":" + config.Password +
		"@tcp(" + config.Host + ":" + strconv.Itoa(config.Port) + ")/" + config.Database +
		"?" + params.Encode()
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn, // DSN data source name
		//DefaultStringSize:         256,   // string 类型字段的默认长度
		//DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		//DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		//DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		//SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
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
