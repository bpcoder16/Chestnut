package cronserver

import (
	"context"
	"strconv"
	"time"

	"github.com/bpcoder16/Chestnut/v4/logit"
	"github.com/bpcoder16/Chestnut/v4/modules/concurrency"
)

type Base struct {
	name               string
	deadLockExpireTime time.Duration
	maxConcurrencyCnt  int

	baseTaskList       []func(context.Context)
	processAddTaskList []func(context.Context)
}

func (b *Base) Before(name string, maxConcurrencyCnt int) {
	b.name = name
	b.maxConcurrencyCnt = maxConcurrencyCnt
	b.processAddTaskList = make([]func(context.Context), 0, 100)
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
	if r := recover(); r != nil {
		logit.Context(ctx).ErrorW(b.name+".Err", r)
	}
}

func (b *Base) GetIsRun(context.Context) bool {
	return true
}

func (b *Base) taskPoolRun(ctx context.Context, taskList []func(context.Context)) {
	if len(taskList) == 0 {
		return
	}
	taskMap := make(map[string]concurrency.Task, len(taskList))
	for index, item := range taskList {
		f := item
		taskMap[strconv.Itoa(index)] = func(ctx context.Context) (any, error) {
			f(ctx)
			return nil, nil
		}
	}
	_, _ = concurrency.RunNamed(ctx, taskMap, concurrency.WithLogField(b.name), concurrency.WithLimit(b.maxConcurrencyCnt))
}
