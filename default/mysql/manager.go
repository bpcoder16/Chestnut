package mysql

import (
	"github.com/bpcoder16/Chestnut/v2/contrib/orm/mysql"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"gorm.io/gorm"
)

var defaultManager *mysql.Manager

func SetManager(configPath string, logger *log.Helper) {
	defaultManager = mysql.NewManager(configPath, logger)
}

func MasterDB() *gorm.DB {
	return defaultManager.MasterDB()
}

func SlaveDB() *gorm.DB {
	return defaultManager.SlaveDB()
}
