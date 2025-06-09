package cron

import (
	"context"
	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/bpcoder16/Chestnut/v2/modules/concurrency"
	"github.com/bpcoder16/Chestnut/v2/modules/lock/nonblock"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Base struct {
	RedisClient *redis.Client

	LockName           string
	name               string
	deadLockExpireTime time.Duration
	maxConcurrencyCnt  int

	baseTaskList       []func(context.Context)
	processAddTaskList []func(context.Context)
}

func (b *Base) Before(name, lockName string, deadLockExpireTime time.Duration, maxConcurrencyCnt int) {
	b.name = name
	b.LockName = lockName
	b.deadLockExpireTime = deadLockExpireTime
	b.maxConcurrencyCnt = maxConcurrencyCnt
	b.processAddTaskList = make([]func(context.Context), 0, 100)
	b.baseTaskList = make([]func(context.Context), 0, 100)
}

func (b *Base) AddBaseTaskList(task func(context.Context)) {
	b.baseTaskList = append(b.baseTaskList, task)
}

func (b *Base) AddProcessAddTaskList(task func(context.Context)) {
	b.processAddTaskList = append(b.processAddTaskList, task)
}

func (b *Base) Init(_ Interface) {
	b.baseTaskList = make([]func(context.Context), 0, 100)
}

func (b *Base) Process(context.Context) {}

func (b *Base) Run(ctx context.Context) {
	b.taskPoolRun(ctx, append(b.baseTaskList, b.processAddTaskList...))
}

func (b *Base) Defer(ctx context.Context) {
	defer nonblock.RedisUnlock(ctx, b.RedisClient, b.LockName)
	if r := recover(); r != nil {
		logit.Context(ctx).ErrorW(b.name+".Err", r)
	}
}

func (b *Base) GetIsRun(ctx context.Context) bool {
	return nonblock.RedisLock(ctx, b.RedisClient, b.LockName, b.deadLockExpireTime)
}

func (b *Base) taskPoolRun(ctx context.Context, taskList []func(context.Context)) {
	if len(taskList) == 0 {
		return
	}
	taskMap := make(map[string]func(ctx context.Context) concurrency.ChanResult)
	if len(taskList) > b.maxConcurrencyCnt {
		cnt := 0
		for index, item := range taskList {
			if cnt >= b.maxConcurrencyCnt {
				_, _ = concurrency.Manager(ctx, taskMap, b.name)
				cnt = 0
				taskMap = make(map[string]func(ctx context.Context) concurrency.ChanResult)
			}
			f := item
			taskMap[strconv.Itoa(index)] = func(ctx context.Context) concurrency.ChanResult {
				f(ctx)
				return concurrency.ChanResult{}
			}
			cnt++
		}
		if len(taskMap) > 0 {
			_, _ = concurrency.Manager(ctx, taskMap, b.name)
		}
	} else {
		for index, item := range taskList {
			f := item
			taskMap[strconv.Itoa(index)] = func(ctx context.Context) concurrency.ChanResult {
				f(ctx)
				return concurrency.ChanResult{}
			}
		}
		_, _ = concurrency.Manager(ctx, taskMap, b.name)
	}
}
