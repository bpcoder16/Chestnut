package bootstrap

import (
	"context"

	"github.com/bpcoder16/Chestnut/v2/appconfig"
	"github.com/bpcoder16/Chestnut/v2/core/asynctask"
)

func Start(ctx context.Context, config *appconfig.AppConfig, goFunc func(f func() error)) {
	if config.AsyncService.Support &&
		config.AsyncService.QueueSize > config.AsyncService.ConsumerSize &&
		config.AsyncService.ConsumerSize > 0 {
		asynctask.StartConsumerPool(
			ctx,
			config.AsyncService.QueueSize,
			config.AsyncService.ConsumerSize,
			config.AsyncService.TaskMaxRetryCnt,
			goFunc,
		)
	}
}
