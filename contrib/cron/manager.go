package cron

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

var (
	scheduler gocron.Scheduler
	once      sync.Once
	mu        sync.Mutex
)

func lazyInit() {
	once.Do(func() {
		var err error
		// 创建一个新的调度器
		scheduler, err = gocron.NewScheduler(
			gocron.WithLocation(env.TimeLocation()),
		)
		if err != nil {
			panic("Create CronScheduler:" + err.Error())
		}
	})
}

func Jobs() []gocron.Job {
	lazyInit()
	mu.Lock()
	defer mu.Unlock()
	return scheduler.Jobs()
}

func NewJob(jobDefinition gocron.JobDefinition, task gocron.Task, jobOptions ...gocron.JobOption) (gocron.Job, error) {
	lazyInit()
	mu.Lock()
	defer mu.Unlock()
	return scheduler.NewJob(jobDefinition, task, jobOptions...)
}

func RemoveJob(uuidStr uuid.UUID) error {
	lazyInit()
	mu.Lock()
	defer mu.Unlock()
	return scheduler.RemoveJob(uuidStr)
}

func Update(uuidStr uuid.UUID, jobDefinition gocron.JobDefinition, task gocron.Task, jobOptions ...gocron.JobOption) (gocron.Job, error) {
	lazyInit()
	mu.Lock()
	defer mu.Unlock()
	return scheduler.Update(uuidStr, jobDefinition, task, jobOptions...)
}

func Run(ctx context.Context) {
	lazyInit()

	scheduler.Start()
	logit.Context(ctx).InfoW("cron.Manager.Run", "CronScheduler started")

	// 捕获系统信号以优雅地关闭调度器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case <-ctx.Done():
		logit.Context(ctx).InfoW("cron.Manager.Run", "Context cancelled, CronScheduler preparing to shutdown")
	case sig := <-sigChan:
		logit.Context(ctx).InfoW("cron.Manager.Run", fmt.Sprintf("Received signal: %v, CronScheduler preparing to shutdown", sig))
	}

	logit.Context(ctx).InfoW("cron.Manager.Run", "CronScheduler shutdown...")
	if err := scheduler.Shutdown(); err != nil {
		logit.Context(ctx).ErrorW("cron.Manager.Run", "CronScheduler shutdown failed:"+err.Error())
	}
	logit.Context(ctx).InfoW("cron.Manager.Run", "CronScheduler shutdown completed, exited")
}
