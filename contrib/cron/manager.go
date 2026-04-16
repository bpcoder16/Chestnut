package cron

import (
	"context"
	"sync"

	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/logit"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

var (
	scheduler        gocron.Scheduler
	once             sync.Once
	mu               sync.Mutex
	schedulerOptions []gocron.SchedulerOption
)

// SetSchedulerOptions 在第一个 Job 注册前调用，将额外选项注入 scheduler（如分布式锁）。
func SetSchedulerOptions(options ...gocron.SchedulerOption) {
	schedulerOptions = append(schedulerOptions, options...)
}

func lazyInit() {
	once.Do(func() {
		var err error
		opts := make([]gocron.SchedulerOption, 0, len(schedulerOptions)+1)
		opts = append(opts, gocron.WithLocation(env.TimeLocation()))
		opts = append(opts, schedulerOptions...)
		scheduler, err = gocron.NewScheduler(opts...)
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

	<-ctx.Done()
	logit.Context(ctx).InfoW("cron.Manager.Run", "Context cancelled, preparing to shutdown")

	if err := scheduler.Shutdown(); err != nil {
		logit.Context(ctx).ErrorW("cron.Manager.Run", "shutdown failed", "error", err.Error())
		return
	}
	logit.Context(ctx).InfoW("cron.Manager.Run", "shutdown completed successfully")
}
