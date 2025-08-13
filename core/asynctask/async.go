package asynctask

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"github.com/bpcoder16/Chestnut/v2/logit"
)

var once sync.Once

type taskData struct {
	f      func(context.Context) error
	errMsg string
	cnt    int
	logId  any
}

var fChan chan taskData

const defaultQueueSize = 10000

func Init(queueSize int) {
	once.Do(func() {
		if queueSize <= defaultQueueSize {
			fChan = make(chan taskData, defaultQueueSize)
		} else {
			fChan = make(chan taskData, queueSize)
		}
	})
}

func AddQueue(ctx context.Context, f func(context.Context) error, errMsg string) {
	Init(defaultQueueSize)
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

func StartConsumerPool(ctx context.Context, consumerCount int, goFunc func(f func() error)) {
	for i := 0; i < consumerCount; i++ {
		goFunc(func() error {
			return Consumer(ctx)
		})
	}
}

func Consumer(ctx context.Context) error {
	Init(defaultQueueSize)
	// 捕获系统信号以优雅地关闭调度器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	ctx = context.WithValue(ctx, log.DefaultMessageKey, "AsyncTask")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigChan:
			return errors.New(fmt.Sprintf("Received signal: %v, Consumer shutdown", sig))
		case f := <-fChan:
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
		if t.cnt >= 3 {
			logit.Context(ctx).ErrorW("async.task", t.errMsg, "async.task.err", err, "cnt", t.cnt)
			return
		}
		logit.Context(ctx).WarnW("async.task", t.errMsg, "async.task.err", err, "cnt", t.cnt)
		fChan <- t
	}
}
