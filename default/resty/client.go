package resty

import (
	"context"
	"time"

	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"github.com/go-resty/resty/v2"
)

var client *resty.Client

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
