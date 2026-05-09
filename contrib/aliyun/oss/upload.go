package oss

import (
	"context"
	"errors"
	"io"

	v2oss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

func PutObjectByScene(ctx context.Context, scene, objectKey string, reader io.Reader) error {
	entry, ok := DefaultManager.scenes[scene]
	if !ok {
		return errors.New("oss: scene not found: " + scene)
	}
	_, err := entry.client.PutObject(ctx, &v2oss.PutObjectRequest{
		Bucket: v2oss.Ptr(entry.sceneConfig.BucketName),
		Key:    v2oss.Ptr(objectKey),
		Body:   reader,
	})
	if err != nil {
		return err
	}
	return nil
}

func SignURLByScene(ctx context.Context, scene, ossPath string, expiredInSec int64) (*v2oss.PresignResult, error) {
	return DefaultManager.SignGetObjectURL(ctx, scene, ossPath, expiredInSec)
}

func ProcessObjectSaveAsByScene(ctx context.Context, scene, sourceObjectKey, targetObjectKey, process string) (*v2oss.ProcessObjectResult, error) {
	return DefaultManager.ProcessObjectSaveAs(ctx, scene, sourceObjectKey, targetObjectKey, process)
}

func ProcessTextWatermarkSaveAsByScene(ctx context.Context, scene, sourceObjectKey string, opts TextWatermarkOptions) (*v2oss.ProcessObjectResult, error) {
	return DefaultManager.ProcessTextWatermarkSaveAs(ctx, scene, sourceObjectKey, opts)
}

func GetObjectFormatByScene(ctx context.Context, scene, objectKey string) (*ObjectFormat, error) {
	return DefaultManager.GetObjectFormat(ctx, scene, objectKey)
}
