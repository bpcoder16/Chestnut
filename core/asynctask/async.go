package asynctask

import (
	"context"
	"errors"
	"sync"

	"github.com/bpcoder16/Chestnut/v3/core/log"
	"github.com/bpcoder16/Chestnut/v3/core/utils"
	"github.com/bpcoder16/Chestnut/v3/logit"
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
	consumerWg       sync.WaitGroup
	shutdownCh       chan struct{}
	shutdownOnce     sync.Once
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
		shutdownCh = make(chan struct{})
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
		consumerWg.Add(1)
		goFunc(func() error {
			defer consumerWg.Done()
			return consumer(ctx)
		})
	}

	// 监听 ctx.Done(),触发优雅关闭
	goFunc(func() error {
		<-ctx.Done()
		GracefulShutdown()
		return ctx.Err()
	})
}

func consumer(ctx context.Context) error {
	ctx = context.WithValue(ctx, log.DefaultMessageKey, "AsyncTask")
	for {
		select {
		case <-shutdownCh:
			// 优雅关闭:继续处理队列中剩余的任务
			for {
				select {
				case f, ok := <-fChan:
					if !ok {
						return errors.New("consumer channel closed")
					}
					task(ctx, f)
				default:
					// 队列已空,退出
					return nil
				}
			}
		case f, ok := <-fChan:
			if !ok {
				return errors.New("consumer channel closed")
			}
			task(ctx, f)
		}
	}
}

// GracefulShutdown 优雅关闭:停止接收新任务,等待所有队列任务消费完成
func GracefulShutdown() {
	shutdownOnce.Do(func() {
		// 关闭 shutdown 通道,通知所有消费者开始清空队列
		close(shutdownCh)
		// 等待所有消费者处理完队列中的任务
		consumerWg.Wait()
		// 关闭任务队列
		close(fChan)
	})
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
