package esclientv7

import (
	"github.com/bpcoder16/Chestnut/v2/core/utils"
)

type Config struct {
	Addresses   []string // A list of Elasticsearch nodes to use.
	Username    string   // Username for HTTP Basic Authentication.
	Password    string   // Password for HTTP Basic Authentication.
	EnableDebug bool
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load elasticsearch conf err:" + err.Error())
	}
	return &config
}
