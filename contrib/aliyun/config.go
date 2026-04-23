package aliyun

import (
	"github.com/bpcoder16/Chestnut/v4/core/utils"
)

type SceneConfig struct {
	AccessKeyId     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	Endpoint        string `yaml:"endpoint"`
	BucketName      string `yaml:"bucketName"`
	TargetDir       string `yaml:"targetDir"`
	StsRoleArn      string `yaml:"stsRoleArn"`
	StsSessionName  string `yaml:"stsSessionName"`
	Region          string `yaml:"region"`
}

type OSSConfig struct {
	Scenes map[string]*SceneConfig `yaml:"scenes"`
}

func LoadOSSConfig(configPath string) *OSSConfig {
	var config OSSConfig
	if err := utils.ParseFile(configPath, &config); err != nil {
		panic("load OSS conf err: " + err.Error())
	}
	return &config
}
