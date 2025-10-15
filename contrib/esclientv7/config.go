package esclientv7

import (
	"github.com/bpcoder16/Chestnut/v3/core/utils"
)

type Config struct {
	Addresses []string // A list of Elasticsearch nodes to use.
	Username  string   // Username for HTTP Basic Authentication.
	Password  string   // Password for HTTP Basic Authentication.
	Log       Log
}

type Log struct {
	Enabled             bool
	RequestBodyEnabled  bool
	ResponseBodyEnabled bool
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseFile(configPath, &config)
	if err != nil {
		panic("load elasticsearch conf err:" + err.Error())
	}
	return &config
}
