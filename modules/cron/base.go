package cron

import (
	"context"
	"github.com/bpcoder16/Chestnut/core/lock/nonblock"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/google/uuid"
	"time"
)

type Base struct {
	Ctx context.Context

	lockName           string
	isRun              bool
	name               string
	deadLockExpireTime time.Duration
	maxConcurrencyCnt  int

	baseTaskList       []func()
	processAddTaskList []func()
}

func (b *Base) Before(name, lockName string, deadLockExpireTime time.Duration, maxConcurrencyCnt int) {
	b.Ctx = context.WithValue(
		context.WithValue(
			context.Background(),
			log.DefaultMessageKey,
			"Cron",
		),
		log.DefaultLogIdKey,
		uuid.New().String(),
	)
	b.name = name
	b.lockName = lockName
	b.deadLockExpireTime = deadLockExpireTime
	b.maxConcurrencyCnt = maxConcurrencyCnt
	b.processAddTaskList = make([]func(), 0, 100)
	b.baseTaskList = make([]func(), 0, 100)
}

func (b *Base) AddBaseTaskList(task func()) {
	b.baseTaskList = append(b.baseTaskList, task)
}

func (b *Base) AddProcessAddTaskList(task func()) {
	b.processAddTaskList = append(b.processAddTaskList, task)
}

func (b *Base) Init(_ Interface) {
	b.baseTaskList = make([]func(), 0, 100)
}

func (b *Base) Process() {}

func (b *Base) Run() {
	b.taskPoolRun(append(b.baseTaskList, b.processAddTaskList...))
}

func (b *Base) Defer() {
	defer nonblock.RedisUnlock(b.Ctx, b.lockName)
	if r := recover(); r != nil {
		logit.Context(b.Ctx).ErrorW(b.name+".Err", r)
	} else {
		if b.isRun {
			logit.Context(b.Ctx).DebugW(b.name+".Status", "Run")
		} else {
			logit.Context(b.Ctx).DebugW(b.name+".Status", "NotRun")
		}
	}
}

func (b *Base) GetIsRun() bool {
	b.isRun = nonblock.RedisLock(b.Ctx, b.lockName, b.deadLockExpireTime)
	return b.isRun
}

func (b *Base) taskPoolRun(taskList []func()) {
	if len(taskList) == 0 {
		return
	}
	for _, task := range taskList {
		task()
	}
}
