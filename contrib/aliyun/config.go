package aliyun

import (
	"github.com/bpcoder16/Chestnut/core/utils"
	"sync"
)

var (
	once   sync.Once
	Config *ConfigItem
)

func InitAliyun(configPath string) {
	once.Do(func() {
		Config = loadAliyunConfig(configPath)
	})
}

type ConfigItem struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	BucketName      string
}

func loadAliyunConfig(configPath string) *ConfigItem {
	var config ConfigItem
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load Aliyun conf err:" + err.Error())
	}
	return &config
}
