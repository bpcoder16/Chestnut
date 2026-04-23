package aliyunoss

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"

	"github.com/bpcoder16/Chestnut/v4/contrib/aliyun/oss"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/default/resty"
	"github.com/bpcoder16/Chestnut/v4/logit"
	goResty "github.com/go-resty/resty/v2"
)

const uploadRetryCnt = 3

func ImageTransferByScene(ctx context.Context, originURL, scene, targetOSSPath string) (err error) {
	var resp *goResty.Response
	resp, err = resty.Client().R().Get(originURL)
	if err != nil {
		logit.Context(ctx).WarnW("aliyunOSS.ImageTransferByScene", "http get failed", "err", err, "url", originURL)
		return
	}
	if resp.IsError() {
		err = errors.New("HTTPStatus:" + resp.Status())
		logit.Context(ctx).WarnW("aliyunOSS.ImageTransferByScene", "http error", "err", err, "url", originURL)
		return
	}

	imageData := bytes.NewReader(resp.Body())
	for i := 0; i < uploadRetryCnt; i++ {
		_, err = imageData.Seek(0, 0)
		if err != nil {
			return
		}
		err = oss.PutObjectByScene(scene, targetOSSPath, imageData)
		if err == nil {
			break
		}
	}
	return
}

func BuildTargetOSSPath(targetDir, originURL string) string {
	ext := filepath.Ext(originURL)
	return filepath.Join(targetDir, utils.UniqueID()+ext)
}
