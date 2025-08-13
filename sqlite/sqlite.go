package sqlite

import (
	"github.com/bpcoder16/Chestnut/v2/contrib/orm/sqlite"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"gorm.io/gorm"
)

var defaultManager *sqlite.Manager

func SetManager(configPath string, logger *log.Helper) {
	defaultManager = sqlite.NewManager(configPath, logger)
}

func DB() *gorm.DB {
	return defaultManager.DB()
}
