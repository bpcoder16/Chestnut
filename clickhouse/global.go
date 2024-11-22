package clickhouse

import (
	"github.com/bpcoder16/Chestnut/contrib/orm/clickhouse"
	"github.com/bpcoder16/Chestnut/core/log"
	"gorm.io/gorm"
)

var defaultManager *clickhouse.Manager

func SetManager(configPath string, logger *log.Helper) {
	defaultManager = clickhouse.NewManager(configPath, logger)
}

func MasterDB() *gorm.DB {
	return defaultManager.MasterDB()
}

func SlaveDB() *gorm.DB {
	return defaultManager.SlaveDB()
}
