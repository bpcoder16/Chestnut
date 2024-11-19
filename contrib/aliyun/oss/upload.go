package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bpcoder16/Chestnut/contrib/aliyun"
	"github.com/google/uuid"
	"mime/multipart"
	"path/filepath"
)

var bucket *oss.Bucket

func InitAliyunOSS(configPath string) {
	aliyun.InitAliyun(configPath)

	// 创建 OSS 客户端
	client, err := oss.New(aliyun.Config.Endpoint, aliyun.Config.AccessKeyId, aliyun.Config.AccessKeySecret)
	if err != nil {
		panic("failed to create OSS client: " + err.Error())
	}

	// 获取存储桶
	bucket, err = client.Bucket(aliyun.Config.BucketName)
	if err != nil {
		panic("failed to get bucket: " + err.Error())
	}
}

func SimpleUpload(fileHeader *multipart.FileHeader, targetDir string) (ossPath string, err error) {
	// 打开上传的文件
	var srcFile multipart.File
	srcFile, err = fileHeader.Open()
	if err != nil {
		return
	}
	defer func(srcFile multipart.File) {
		_ = srcFile.Close()
	}(srcFile)

	// 获取文件的扩展名
	ext := filepath.Ext(fileHeader.Filename)
	ossPath = filepath.Join(targetDir, uuid.New().String()+ext)

	err = bucket.PutObject(ossPath, srcFile)
	if err != nil {
		return
	}
	return
}
