package resty

import (
	"context"
	"errors"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/go-resty/resty/v2"
)

var client *resty.Client

var (
	ErrClientNotInitialized     = errors.New("resty client is not initialized")
	ErrMissingMicroServiceLogID = errors.New("missing " + log.MicroServiceLogIdHeader)
)

func SetClient(logger *log.Helper) {
	client = resty.New()

	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		ctx := resp.Request.Context()
		ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "HTTPClient")
		// 计算访问耗时
		elapsed := time.Since(resp.Request.Time)

		logger.WithContext(ctx).DebugW(
			"URL", resp.Request.URL,
			"requestBody", resp.Request.Body,
			"respHTTPStatus", resp.StatusCode(),
			"respBody", resp.String(),
			"costTime", utils.ShowDurationString(elapsed),
		)
		return nil
	})

	// TODO 临时设置全局的超时时间，需要等到后续 resty 升级到 v3 后，支持单个请求的超时设置
	client.SetTimeout(15 * time.Second)
}

func Client() *resty.Client {
	return client
}

// MicroServiceRequest 创建微服务间 HTTP 请求，并自动透传当前请求的 logId。
func MicroServiceRequest(ctx context.Context) (*resty.Request, error) {
	if client == nil {
		return nil, ErrClientNotInitialized
	}
	logId := log.LogIdFromContext(ctx)
	if logId == "" {
		return nil, ErrMissingMicroServiceLogID
	}
	return client.R().
		SetContext(ctx).
		SetHeader(log.MicroServiceLogIdHeader, logId), nil
}
