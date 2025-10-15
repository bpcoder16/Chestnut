package sqlite

import (
	"github.com/bpcoder16/Chestnut/v3/contrib/orm/sqlite"
	"github.com/bpcoder16/Chestnut/v3/core/log"
	"gorm.io/gorm"
)

var defaultManager *sqlite.Manager

func SetManager(configPath string, logger *log.Helper) {
	defaultManager = sqlite.NewManager(configPath, logger)
}

func DefaultClient() *gorm.DB {
	return defaultManager.DB()
}
