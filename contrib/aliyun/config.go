package aliyun

import (
	"github.com/bpcoder16/Chestnut/v4/core/utils"
)

type SceneConfig struct {
	BucketType      string   `yaml:"bucketType"`
	AccessKeyId     string   `yaml:"accessKeyId"`
	AccessKeySecret string   `yaml:"accessKeySecret"`
	Endpoint        string   `yaml:"endpoint"`
	StsEndpoint     string   `yaml:"stsEndpoint"`
	BucketName      string   `yaml:"bucketName"`
	CdnBaseURL      string   `yaml:"cdnBaseURL"`
	StsRoleArn      string   `yaml:"stsRoleArn"`
	StsSessionName  string   `yaml:"stsSessionName"`
	Region          string   `yaml:"region"`
	TargetDir       string   `yaml:"targetDir"`
	MaxBatchCount   int      `yaml:"maxBatchCount"`
	MaxSize         int64    `yaml:"maxSize"`
	AllowedExts     []string `yaml:"allowedExts"`
}

type OSSBucketConfig struct {
	AccessKeyId     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	Endpoint        string `yaml:"endpoint"`
	StsEndpoint     string `yaml:"stsEndpoint"`
	BucketName      string `yaml:"bucketName"`
	CdnBaseURL      string `yaml:"cdnBaseURL"`
	StsRoleArn      string `yaml:"stsRoleArn"`
	StsSessionName  string `yaml:"stsSessionName"`
	Region          string `yaml:"region"`
}

type OSSConfig struct {
	Buckets map[string]*OSSBucketConfig `yaml:"buckets"`
	Scenes  map[string]*SceneConfig     `yaml:"scenes"`
}

func LoadOSSConfig(configPath string) *OSSConfig {
	var config OSSConfig
	if err := utils.ParseFile(configPath, &config); err != nil {
		panic("load OSS conf err: " + err.Error())
	}
	return &config
}
