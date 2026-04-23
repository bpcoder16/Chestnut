package oss

import (
	"io"

	gosdk "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func PutObjectByScene(scene, objectKey string, reader io.Reader) error {
	bucket, err := DefaultManager.GetBucket(scene)
	if err != nil {
		return err
	}
	return bucket.PutObject(objectKey, reader)
}

func SignURLByScene(scene, ossPath string, expiredInSec int64) (string, error) {
	bucket, err := DefaultManager.GetBucket(scene)
	if err != nil {
		return "", err
	}
	return bucket.SignURL(ossPath, gosdk.HTTPGet, expiredInSec)
}
