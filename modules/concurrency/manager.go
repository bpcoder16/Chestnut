package concurrency

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/bpcoder16/Chestnut/v4/core/gtask"
	corelog "github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/logit"
)

const (
	defaultLogField   = "default"
	panicStackBufSize = 64 * 1024
)

type Task func(ctx context.Context) (any, error)

type Result struct {
	Value any
	Err   error
}

type PanicError struct {
	TaskName string
	Value    any
	Stack    []byte
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("concurrency task %s panic: %v", e.TaskName, e.Value)
}

type PanicHandler func(ctx context.Context, taskName string, recovered any, stack []byte)

type Option func(*options)

type options struct {
	logField     string
	limit        int
	failFast     bool
	recoverPanic bool
	panicHandler PanicHandler
}

func defaultOptions() options {
	return options{
		logField:     defaultLogField,
		recoverPanic: true,
		panicHandler: defaultPanicHandler,
	}
}

func WithLogField(logField string) Option {
	return func(o *options) {
		if logField != "" {
			o.logField = logField
		}
	}
}

func WithLimit(limit int) Option {
	return func(o *options) {
		o.limit = limit
	}
}

func WithFailFast(enabled bool) Option {
	return func(o *options) {
		o.failFast = enabled
	}
}

func WithRecoverPanic(enabled bool) Option {
	return func(o *options) {
		o.recoverPanic = enabled
	}
}

func WithPanicHandler(handler PanicHandler) Option {
	return func(o *options) {
		if handler != nil {
			o.panicHandler = handler
		}
	}
}

func RunNamed(ctx context.Context, tasks map[string]Task, opts ...Option) (map[string]Result, error) {
	o := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	defer utils.TimeCostLog(ctx, "concurrency.RunNamed."+o.logField)()
	results := make(map[string]Result, len(tasks))
	if len(tasks) == 0 {
		return results, nil
	}

	g, groupCtx := gtask.WithContext(ctx)
	if o.limit > 0 {
		g.SetLimit(o.limit)
	}

	var mu sync.Mutex
	errList := make([]error, 0)
	for name, task := range tasks {
		taskName := name
		taskFunc := task
		g.Go(func() error {
			result, taskErr := runTask(groupCtx, taskName, taskFunc, o)

			mu.Lock()
			results[taskName] = result
			if result.Err != nil {
				errList = append(errList, result.Err)
			}
			mu.Unlock()

			if o.failFast {
				return taskErr
			}
			return nil
		})
	}

	_ = g.Wait()

	err := errors.Join(errList...)
	return results, err
}

func runTask(ctx context.Context, name string, task Task, o options) (result Result, err error) {
	ctx = context.WithValue(ctx, corelog.DefaultConcurrencyLogIdKey, utils.UniqueID())
	defer utils.TimeCostLog(ctx, "concurrency.RunNamed."+o.logField+"."+name)()

	if o.recoverPanic {
		defer func() {
			if recovered := recover(); recovered != nil {
				stack := make([]byte, panicStackBufSize)
				stack = stack[:runtime.Stack(stack, false)]
				panicErr := &PanicError{
					TaskName: name,
					Value:    recovered,
					Stack:    stack,
				}
				if o.panicHandler != nil {
					o.panicHandler(ctx, name, recovered, stack)
				}
				result.Err = panicErr
				err = panicErr
			}
		}()
	}

	value, taskErr := task(ctx)
	result = Result{
		Value: value,
		Err:   taskErr,
	}
	return result, taskErr
}

func defaultPanicHandler(ctx context.Context, taskName string, recovered any, stack []byte) {
	logit.Context(ctx).ErrorW(
		"concurrency.RunNamed.Panic", "并发任务 panic",
		"taskName", taskName,
		"panic", recovered,
		"stack", string(stack),
	)
}
