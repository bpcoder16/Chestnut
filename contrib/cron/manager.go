package cron

import (
	"context"
	"sync"

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

	select {
	case <-ctx.Done():
		logit.Context(ctx).InfoW("cron.Manager.Run", "Context cancelled, CronScheduler preparing to shutdown")
	}

	logit.Context(ctx).InfoW("cron.Manager.Run", "CronScheduler shutdown...")
	if err := scheduler.Shutdown(); err != nil {
		logit.Context(ctx).ErrorW("cron.Manager.Run", "CronScheduler shutdown failed:"+err.Error())
	}
	logit.Context(ctx).InfoW("cron.Manager.Run", "CronScheduler shutdown completed, exited")
}
