package oss

import "github.com/aliyun/aliyun-oss-go-sdk/oss"

func SignURL(ossPath string, expiredInSec int64) (signedURL string, err error) {
	return Bucket.SignURL(ossPath, oss.HTTPGet, expiredInSec)
}
