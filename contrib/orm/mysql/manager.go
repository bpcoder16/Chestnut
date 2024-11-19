package mysql

import (
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/core/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strconv"
	"time"
)

type GormDBManager struct {
	dbMaster *gorm.DB
	dbSlaves []*gorm.DB
	logger   *log.Helper
	config   *Config
}

func NewGormDBManager(configPath string, logger *log.Helper) *GormDBManager {
	manager := &GormDBManager{
		logger:   logger,
		config:   loadMySQLConfig(configPath),
		dbMaster: nil,
		dbSlaves: make([]*gorm.DB, 0, 10),
	}
	manager.connectMaster()
	manager.connectSlaves()
	return manager
}

func (g *GormDBManager) MasterDB() *gorm.DB {
	return g.dbMaster
}

func (g *GormDBManager) SlaveDB() *gorm.DB {
	switch len(g.dbSlaves) {
	case 0:
		return g.MasterDB()
	case 1:
		return g.dbSlaves[0]
	default:
		return g.dbSlaves[utils.RandIntN(len(g.dbSlaves))]
	}
}

func (g *GormDBManager) connect(config *ConfigItem) *gorm.DB {
	dsn := config.Username + ":" + config.Password +
		"@tcp(" + config.Host + ":" + strconv.Itoa(config.Port) + ")/" + config.Database +
		"?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn, // DSN data source name
		//DefaultStringSize:         256,   // string 类型字段的默认长度
		//DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		//DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		//DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		//SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		Logger: NewLogger(g.logger, logger.Config{
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

func (g *GormDBManager) setConnectionPool(db *gorm.DB, config *ConfigItem) {
	sqlDB, _ := db.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)
}

func (g *GormDBManager) connectMaster() {
	g.dbMaster = g.connect(g.config.Master)
	g.setConnectionPool(g.dbMaster, g.config.Master)
}

func (g *GormDBManager) connectSlaves() {
	if len(g.config.Slaves) > 0 {
		for _, slaveConfig := range g.config.Slaves {
			db := g.connect(slaveConfig)
			g.setConnectionPool(db, slaveConfig)
			g.dbSlaves = append(g.dbSlaves, db)
		}
	}
}
