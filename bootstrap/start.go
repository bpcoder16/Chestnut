package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/bpcoder16/Chestnut/v4/appconfig"
	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/core/asynctask"
	"github.com/bpcoder16/Chestnut/v4/default/locallock"
	"github.com/bpcoder16/Chestnut/v4/logit"
)

func Start(ctx context.Context, config *appconfig.AppConfig, goFunc func(f func() error)) {
	goFunc(func() error {
		// 捕获系统信号以优雅地关闭调度器
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			logit.Context(ctx).InfoW("goFunc.Run", "Context cancelled, shutdown completed successfully")
			return errors.New("ctx.Done")
		case sig := <-sigChan:
			logit.Context(ctx).InfoW("goFunc.Run", fmt.Sprintf("Received signal: %v, shutdown completed successfully", sig))
			return fmt.Errorf("received signal: %v, shutdown completed successfully", sig)
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
	if config.Default.LocalLockPoolSupport {
		goFunc(func() error {
			locallock.Run(
				ctx,
				path.Join(env.ConfigDirPath(), "local_lock_pool.yaml"),
			)
			return nil
		})
	}
}
