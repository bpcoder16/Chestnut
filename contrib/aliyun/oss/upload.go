package oss

import (
	"context"
	"errors"
	"io"

	v2oss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

func PutObjectByScene(scene, objectKey string, reader io.Reader) error {
	entry, ok := DefaultManager.scenes[scene]
	if !ok {
		return errors.New("oss: scene not found: " + scene)
	}
	_, err := entry.client.PutObject(context.Background(), &v2oss.PutObjectRequest{
		Bucket: v2oss.Ptr(entry.bucketConfig.BucketName),
		Key:    v2oss.Ptr(objectKey),
		Body:   reader,
	})
	if err != nil {
		return err
	}
	return nil
}

func SignURLByScene(scene, ossPath string, expiredInSec int64) (*v2oss.PresignResult, error) {
	return DefaultManager.SignGetObjectURL(context.Background(), scene, ossPath, expiredInSec)
}
