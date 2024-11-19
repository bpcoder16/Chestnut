package aliyunoss

import (
	"bytes"
	"context"
	"errors"
	"github.com/bpcoder16/Chestnut/contrib/aliyun/oss"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/bpcoder16/Chestnut/resty"
	goResty "github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"mime/multipart"
	"path/filepath"
)

const (
	uploadRetryCnt = 3
)

// ImageTransfer 图片转移
func ImageTransfer(ctx context.Context, originURL, targetOSSPath string) (err error) {
	var resp *goResty.Response
	resp, err = resty.Client().R().Get(originURL)
	if err != nil {
		logit.Context(ctx).WarnW("aliyunOSS.ImageTransfer.Err:", err)
		return
	}
	if resp.IsError() {
		err = errors.New("HTTPStatus:" + resp.Status())
		logit.Context(ctx).WarnW("aliyunOSS.ImageTransfer.Err:", err)
		return
	}

	// 转换为 io.Reader 传递给 OSS
	imageData := bytes.NewReader(resp.Body())

	for i := 0; i < uploadRetryCnt; i++ {
		err = oss.PutObject(targetOSSPath, imageData)
		if err == nil {
			break
		}
	}

	return
}

func BuildTargetOSSPath(targetDir, originURL string) string {
	// 获取文件的扩展名
	ext := filepath.Ext(originURL)
	return filepath.Join(targetDir, uuid.New().String()+ext)
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

	ossPath = BuildTargetOSSPath(targetDir, fileHeader.Filename)

	err = oss.PutObject(ossPath, srcFile)
	if err != nil {
		return
	}
	return
}
