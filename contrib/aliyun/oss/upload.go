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

func ProcessObjectSaveAsByScene(scene, sourceObjectKey, targetObjectKey, process string) (*v2oss.ProcessObjectResult, error) {
	return DefaultManager.ProcessObjectSaveAs(context.Background(), scene, sourceObjectKey, targetObjectKey, process)
}

func ProcessTextWatermarkSaveAsByScene(scene, sourceObjectKey string, opts TextWatermarkOptions) (*v2oss.ProcessObjectResult, error) {
	return DefaultManager.ProcessTextWatermarkSaveAs(context.Background(), scene, sourceObjectKey, opts)
}

func GetObjectFormatByScene(scene, objectKey string) (*ObjectFormat, error) {
	return DefaultManager.GetObjectFormat(context.Background(), scene, objectKey)
}
