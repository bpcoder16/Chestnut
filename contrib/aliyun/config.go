package aliyun

import (
	"github.com/bpcoder16/Chestnut/core/utils"
	"sync"
)

var (
	once    sync.Once
	_config *Config
)

type Config struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	Endpoint        string `json:"endpoint"`
	BucketName      string `json:"bucketName"`
}

func InitAliyunConfig(configPath string) *Config {
	once.Do(func() {
		_config = loadConfig(configPath)
	})
	return _config
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load Aliyun conf err:" + err.Error())
	}
	return &config
}
