package mysql

import (
	"github.com/bpcoder16/Chestnut/contrib/orm/mysql"
	"github.com/bpcoder16/Chestnut/core/log"
	"gorm.io/gorm"
)

var defaultMySQLGormDBManager *mysql.GormDBManager

func SetManager(configPath string, logger *log.Helper) {
	defaultMySQLGormDBManager = mysql.NewGormDBManager(configPath, logger)
}

func MasterDB() *gorm.DB {
	return defaultMySQLGormDBManager.MasterDB()
}

func SlaveDB() *gorm.DB {
	return defaultMySQLGormDBManager.SlaveDB()
}
