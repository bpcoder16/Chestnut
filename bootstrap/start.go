package bootstrap

import (
	"context"
	"github.com/bpcoder16/Chestnut/v2/appconfig"
	"github.com/bpcoder16/Chestnut/v2/core/asynctask"
)

func Start(ctx context.Context, config *appconfig.AppConfig, goFunc func(f func() error)) {
	if config.ConsumerSize > 0 {
		asynctask.StartConsumerPool(ctx, config.ConsumerSize, goFunc)
	}
}
