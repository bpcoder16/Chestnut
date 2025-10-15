package locallock

import "github.com/bpcoder16/Chestnut/v3/core/utils"

type Config struct {
	TTLSec   int
	SweepSec int
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load local_lock_pool conf err:" + err.Error())
	}
	return &config
}
