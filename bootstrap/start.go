package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bpcoder16/Chestnut/v2/appconfig"
	"github.com/bpcoder16/Chestnut/v2/core/asynctask"
	"github.com/bpcoder16/Chestnut/v2/logit"
)

func Start(ctx context.Context, config *appconfig.AppConfig, goFunc func(f func() error)) {
	goFunc(func() error {
		// 捕获系统信号以优雅地关闭调度器
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			logit.Context(ctx).InfoW("goFunc.Run", "Context cancelled, goFunc preparing to shutdown")
			return errors.New("ctx.Done")
		case sig := <-sigChan:
			logit.Context(ctx).InfoW("goFunc.Run", fmt.Sprintf("Received signal: %v, goFunc preparing to shutdown", sig))
			return fmt.Errorf("received signal: %v, goFunc preparing to shutdown", sig)
		}
	})
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
