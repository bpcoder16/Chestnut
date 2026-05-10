package concurrency

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

const (
	limitTestTaskCount = 10
	limitTestLimit     = 3
	limitTestSleep     = 10 * time.Millisecond
)

func TestRunNamedEmptyTasks(t *testing.T) {
	results, err := RunNamed(context.Background(), nil, WithPanicHandler(noopPanicHandler))
	if err != nil {
		t.Fatalf("RunNamed() error = %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("RunNamed() results len = %d, want 0", len(results))
	}
}

func TestRunNamedCollectResults(t *testing.T) {
	tasks := map[string]Task{
		"first": func(context.Context) (any, error) {
			return "first-result", nil
		},
		"second": func(context.Context) (any, error) {
			return 2, nil
		},
	}

	results, err := RunNamed(context.Background(), tasks, WithPanicHandler(noopPanicHandler))
	if err != nil {
		t.Fatalf("RunNamed() error = %v", err)
	}
	if results["first"].Value != "first-result" {
		t.Fatalf("first result = %v, want first-result", results["first"].Value)
	}
	if results["second"].Value != 2 {
		t.Fatalf("second result = %v, want 2", results["second"].Value)
	}
}

func TestRunNamedCollectErrors(t *testing.T) {
	taskErr := errors.New("task failed")
	tasks := map[string]Task{
		"failed": func(context.Context) (any, error) {
			return nil, taskErr
		},
		"success": func(context.Context) (any, error) {
			return "ok", nil
		},
	}

	results, err := RunNamed(context.Background(), tasks, WithPanicHandler(noopPanicHandler))
	if !errors.Is(err, taskErr) {
		t.Fatalf("RunNamed() error = %v, want %v", err, taskErr)
	}
	if !errors.Is(results["failed"].Err, taskErr) {
		t.Fatalf("failed result error = %v, want %v", results["failed"].Err, taskErr)
	}
	if results["success"].Value != "ok" {
		t.Fatalf("success result = %v, want ok", results["success"].Value)
	}
}

func TestRunNamedRecoverPanic(t *testing.T) {
	var called atomic.Bool
	tasks := map[string]Task{
		"panic": func(context.Context) (any, error) {
			panic("boom")
		},
	}

	results, err := RunNamed(context.Background(), tasks, WithPanicHandler(func(context.Context, string, any, []byte) {
		called.Store(true)
	}))
	if err == nil {
		t.Fatal("RunNamed() error = nil, want panic error")
	}
	var panicErr *PanicError
	if !errors.As(err, &panicErr) {
		t.Fatalf("RunNamed() error = %T, want *PanicError", err)
	}
	if panicErr.TaskName != "panic" {
		t.Fatalf("panic task name = %s, want panic", panicErr.TaskName)
	}
	if !called.Load() {
		t.Fatal("panic handler was not called")
	}
	if !errors.As(results["panic"].Err, &panicErr) {
		t.Fatalf("panic result error = %T, want *PanicError", results["panic"].Err)
	}
}

func TestRunNamedLimit(t *testing.T) {
	var active atomic.Int32
	var maxActive atomic.Int32
	tasks := make(map[string]Task, limitTestTaskCount)
	for i := 0; i < limitTestTaskCount; i++ {
		tasks[strconv.Itoa(i)] = func(context.Context) (any, error) {
			current := active.Add(1)
			for {
				max := maxActive.Load()
				if current <= max || maxActive.CompareAndSwap(max, current) {
					break
				}
			}
			time.Sleep(limitTestSleep)
			active.Add(-1)
			return nil, nil
		}
	}

	_, err := RunNamed(context.Background(), tasks, WithLimit(limitTestLimit), WithPanicHandler(noopPanicHandler))
	if err != nil {
		t.Fatalf("RunNamed() error = %v", err)
	}
	if maxActive.Load() > limitTestLimit {
		t.Fatalf("max active tasks = %d, want <= %d", maxActive.Load(), limitTestLimit)
	}
}

func noopPanicHandler(context.Context, string, any, []byte) {}
