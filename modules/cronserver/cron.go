package cronserver

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/bpcoder16/Chestnut/v4/contrib/cron"
	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/logit"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
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

func Run(ctx context.Context, configPath string) {
	config := loadConfig(configPath)
	if !config.IsRunCron {
		return
	}

	ctx = context.WithValue(ctx, log.DefaultMessageKey, "Cron")
	for _, cronConfig := range config.CronList {
		cronController, cronErr := getCron(cronConfig)
		if cronErr == nil {
			cronConfigNew := cronConfig
			var jobDefinition gocron.JobDefinition
			switch cronConfigNew.JobType {
			case "CronJob":
				jobDefinition = gocron.CronJob(cronConfigNew.CronJobParams.Crontab, true)
			case "DurationJob":
				jobDefinition = gocron.DurationJob(time.Duration(cronConfigNew.DurationJobParams.EveryMillisecond) * time.Millisecond)
			case "DurationRandomJob":
				jobDefinition = gocron.DurationRandomJob(
					time.Duration(cronConfigNew.DurationRandomJobParams.MinMillisecond)*time.Millisecond,
					time.Duration(cronConfigNew.DurationRandomJobParams.MaxMillisecond)*time.Millisecond,
				)
			default:
				continue
			}
			_, errJob := cron.NewJob(
				jobDefinition,
				gocron.NewTask(func(taskCtx context.Context, task Interface, configItem ConfigItem) {
					taskCtx = context.WithValue(taskCtx, log.DefaultLogIdKey, utils.UniqueID())
					taskCtx = context.WithValue(taskCtx, log.DefaultCronActionKey, configItem.Name)

					task.Before(configItem.Name, configItem.MaxConcurrencyCnt)
					if task.GetIsRun(taskCtx) {
						defer task.Defer(taskCtx)
						task.Process(taskCtx)
						task.Run(taskCtx)
						logit.Context(taskCtx).DebugW(configItem.Name+".Status", "Run")
					} else {
						logit.Context(taskCtx).DebugW(configItem.Name+".Status", "NotRun")
					}
				}, ctx, cronController, cronConfigNew),
				gocron.WithContext(ctx),
				gocron.WithName(cronConfigNew.Name),
				gocron.WithSingletonMode(gocron.LimitModeWait),
				gocron.WithIdentifier(uuid.New()),
			)
			if errJob != nil {
				panic("cronServer.Run.Err:" + errJob.Error())
			}
		}
	}

	cron.Run(ctx)
}
