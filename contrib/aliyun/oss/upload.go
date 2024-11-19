package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bpcoder16/Chestnut/contrib/aliyun"
	"io"
)

var Bucket *oss.Bucket

func InitAliyunOSS(configPath string) {
	aliyun.InitAliyun(configPath)

	// 创建 OSS 客户端
	client, err := oss.New(aliyun.Config.Endpoint, aliyun.Config.AccessKeyId, aliyun.Config.AccessKeySecret)
	if err != nil {
		panic("failed to create OSS client: " + err.Error())
	}

	// 获取存储桶
	Bucket, err = client.Bucket(aliyun.Config.BucketName)
	if err != nil {
		panic("failed to get bucket: " + err.Error())
	}
}

func PutObject(objectKey string, reader io.Reader) error {
	return Bucket.PutObject(objectKey, reader)
}
