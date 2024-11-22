package cron

import (
	"context"
	"errors"
	"fmt"
	"github.com/bpcoder16/Chestnut/modules/appconfig/env"
	"github.com/go-co-op/gocron/v2"
	"reflect"
	"time"
)

var cronMap = make(map[string]Interface)

func RegisterCron(cronName string, cron Interface) {
	cronMap[cronName] = cron
}

func getCron(cronConfig ConfigItem) (cron Interface, err error) {
	if len(cronMap) == 0 {
		err = errors.New("cron config list is empty")
		return
	}
	var exist bool
	var cronTemplate Interface
	cronTemplate, exist = cronMap[cronConfig.Name]
	if !exist {
		err = errors.New("cron config[" + cronConfig.Name + "] is not exist")
		return
	}
	cron, _ = reflect.New(reflect.TypeOf(cronTemplate).Elem()).Interface().(Interface)
	cron.Init(cronTemplate)
	return
}

func Run(ctx context.Context) {
	if !config.IsRunCron {
		return
	}

	s, err := gocron.NewScheduler(
		gocron.WithLocation(env.TimeLocation()),
	)
	if err != nil {
		panic(err)
	}

	for _, cronConfig := range config.CronList {
		cronController, cronErr := getCron(cronConfig)
		if cronErr == nil {
			cronConfigNew := cronConfig
			var job gocron.JobDefinition
			switch cronConfigNew.JobType {
			case "CronJob":
				job = gocron.CronJob(cronConfigNew.CronJobParams.Crontab, cronConfigNew.CronJobParams.WithSeconds)
			case "DurationJob":
				job = gocron.DurationJob(time.Duration(cronConfigNew.DurationJobParams.EveryMillisecond) * time.Millisecond)
			case "DurationRandomJob":
				job = gocron.DurationRandomJob(
					time.Duration(cronConfigNew.DurationRandomJobParams.MinMillisecond)*time.Millisecond,
					time.Duration(cronConfigNew.DurationRandomJobParams.MaxMillisecond)*time.Millisecond,
				)
			default:
				continue
			}
			_, _ = s.NewJob(
				job,
				gocron.NewTask(func(task Interface, configItem ConfigItem, lockPreName string) {
					task.Before(
						configItem.Name,
						env.AppName()+":"+lockPreName+":"+configItem.Name,
						time.Duration(configItem.DeadLockExpireMillisecond)*time.Millisecond,
						configItem.MaxConcurrencyCnt,
					)
					defer task.Defer()
					if task.GetIsRun() {
						task.Process()
						task.Run()
					}
				}, cronController, cronConfigNew, config.LockPreName),
			)
		}
	}
	s.Start()

	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled, cron shutting down...")
	}

	fmt.Println("Stopping cron...")
	_ = s.Shutdown()
	fmt.Println("Cron stopped. Exiting.")
}
