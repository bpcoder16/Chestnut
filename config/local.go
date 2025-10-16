package config

import (
	"path/filepath"

	"github.com/bpcoder16/Chestnut/v3/core/utils"
)

func ParseLocalConfig(confPath string, configPtr *AppConfig) (err error) {
	if confPath, err = filepath.Abs(confPath); err == nil {
		if err = utils.ParseFile(confPath, configPtr); err == nil {
			err = configPtr.Check()
		}
	}
	return
}
