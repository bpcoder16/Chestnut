package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bpcoder16/Chestnut/v2/contrib/aliyun"
)

var Bucket *oss.Bucket

func InitAliyunOSS(configPath string) {
	config := aliyun.InitAliyunConfig(configPath)

	// 创建 OSS 客户端
	client, err := oss.New(config.Endpoint, config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		panic("failed to create OSS client: " + err.Error())
	}

	// 获取存储桶
	Bucket, err = client.Bucket(config.BucketName)
	if err != nil {
		panic("failed to get bucket: " + err.Error())
	}
}
