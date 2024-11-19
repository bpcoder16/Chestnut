package clickhouse

import (
	"github.com/bpcoder16/Chestnut/contrib/orm/clickhouse"
	"github.com/bpcoder16/Chestnut/core/log"
	"gorm.io/gorm"
)

var defaultClickhouseGormDBManager *clickhouse.GormDBManager

func SetManager(configPath string, logger *log.Helper) {
	defaultClickhouseGormDBManager = clickhouse.NewGormDBManager(configPath, logger)
}

func MasterDB() *gorm.DB {
	return defaultClickhouseGormDBManager.MasterDB()
}

func SlaveDB() *gorm.DB {
	return defaultClickhouseGormDBManager.SlaveDB()
}
