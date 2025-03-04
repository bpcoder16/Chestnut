package resty

import (
	"context"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"github.com/go-resty/resty/v2"
	"time"
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
}

func Client() *resty.Client {
	return client
}
