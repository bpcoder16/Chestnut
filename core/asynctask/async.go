package asynctask

import (
	"context"
	"errors"
	"sync"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/logit"
)

type taskData struct {
	f      func(context.Context) error
	errMsg string
	cnt    int
	logId  any
}

var (
	once             sync.Once
	fChan            chan taskData
	fQueueSize       int
	qTaskMaxRetryCnt int
)

func SetQueueSize(queueSize int) {
	fQueueSize = queueSize
}

const (
	defaultQueueSize       = 10000
	defaultTaskMaxRetryCnt = 3
)

func lazyInit() {
	once.Do(func() {
		if fQueueSize > 0 {
			fChan = make(chan taskData, fQueueSize)
		} else {
			fChan = make(chan taskData, defaultQueueSize)
		}
	})
}

func getTaskMaxRetryCnt() int {
	if qTaskMaxRetryCnt <= 0 {
		return defaultTaskMaxRetryCnt
	}
	return qTaskMaxRetryCnt
}

func AddQueue(ctx context.Context, f func(context.Context) error, errMsg string) {
	lazyInit()
	logId := ctx.Value(log.DefaultLogIdKey)
	if logId == nil {
		logId = utils.UniqueID()
	}
	fChan <- taskData{
		f:      f,
		errMsg: errMsg,
		cnt:    0,
		logId:  logId,
	}
}

func StartConsumerPool(ctx context.Context, queueSize, consumerSize, taskMaxRetryCnt int, goFunc func(f func() error)) {
	qTaskMaxRetryCnt = taskMaxRetryCnt
	SetQueueSize(queueSize)
	lazyInit()
	for i := 0; i < consumerSize; i++ {
		goFunc(func() error {
			return consumer(ctx)
		})
	}
}

func consumer(ctx context.Context) error {
	ctx = context.WithValue(ctx, log.DefaultMessageKey, "AsyncTask")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case f, ok := <-fChan:
			if !ok {
				return errors.New("consumer Channel closed")
			}
			task(ctx, f)
		}
	}
}

func task(ctx context.Context, t taskData) {
	ctx = context.WithValue(ctx, log.DefaultLogIdKey, t.logId)
	defer func() {
		if r := recover(); r != nil {
			logit.Context(ctx).ErrorW("async.task", t.errMsg, "async.task.panic", r)
		}
	}()
	if err := t.f(ctx); err != nil {
		t.cnt++
		if t.cnt >= getTaskMaxRetryCnt() {
			logit.Context(ctx).ErrorW("async.task", t.errMsg, "async.task.err", err, "cnt", t.cnt)
			return
		}
		logit.Context(ctx).WarnW("async.task", t.errMsg, "async.task.err", err, "cnt", t.cnt)
		fChan <- t
	}
}
