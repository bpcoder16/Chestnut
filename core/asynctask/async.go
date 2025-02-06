package asynctask

import (
	"context"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"github.com/bpcoder16/Chestnut/v2/logit"
)

type taskData struct {
	f      func(context.Context) error
	errMsg string
	cnt    int
	logId  any
}

var fChan = make(chan taskData, 10000)

func AddQueue(ctx context.Context, f func(context.Context) error, errMsg string) {
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

func Consumer(ctx context.Context) error {
	ctx = context.WithValue(ctx, log.DefaultMessageKey, "AsyncTask")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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
