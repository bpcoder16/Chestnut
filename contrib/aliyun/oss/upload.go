package oss

import (
	"io"
)

func PutObject(objectKey string, reader io.Reader) error {
	return Bucket.PutObject(objectKey, reader)
}
