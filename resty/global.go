package resty

import (
	"github.com/go-resty/resty/v2"
)

var client *resty.Client

func SetClient() {
	client = resty.New()
}

func Client() *resty.Client {
	return client
}
