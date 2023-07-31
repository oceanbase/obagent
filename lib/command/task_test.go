package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

var testStore = NewFileTaskStore(os.TempDir())

func TestTask_Init(t *testing.T) {
	cmd := NewTask(func(*Input, ProgressFunc) (interface{}, error) {
		return "", nil
	}, "", TypeText)
	if cmd.ResponseType() != TypeText || cmd.fn == nil {
		t.Error("init wrong")
	}
}

func TestTask_Execute(t *testing.T) {
	cmd := NewTask(func(input *Input, progressFunc ProgressFunc) (interface{}, error) {
		progressFunc(1)
		return "hello " + input.Param().(string), nil
	}, "", TypeText)
	execCtx := NewInputExecutionContext(context.Background(), "world")
	_ = cmd.Execute(execCtx)
	result, err := execCtx.Output().Get()
	if err != nil || result != "hello world" {
		t.Error("bad result")
	}
	p := execCtx.Output().Progress()
	if p != 1 {
		t.Error("bad progress")
	}
}

func TestTask_ExecutorInit(t *testing.T) {
	executor := NewExecutor(testStore)
	if executor.id == 0 || executor.nowFunc == nil {
		t.Error("init wrong")
	}
}

func TestTask_Executor(t *testing.T) {
	cmd := NewTask(func(input *Input, progressFunc ProgressFunc) (interface{}, error) {
		progressFunc(1)
		time.Sleep(time.Second)
		return "hello " + input.Param().(string), nil
	}, "", TypeText)
	executor := NewExecutor(testStore)
	token, _ := executor.Execute(cmd, NewInput(context.Background(), "world"))
	if token.id == "" {
		t.Error("bad token")
	}
	time.Sleep(time.Millisecond * 100)
	result, ok := executor.GetResult(token)
	if !ok {
		t.Error("find result failed")
		return
	}
	if result.Progress != 1 {
		t.Error("can not get progress")
	}
	e, _ := executor.GetExecution(token)
	if e.Command() == nil || e.ExecutionContext() == nil {
		t.Error("bad execution")
	}
	execCtx, _ := executor.getExecContext(token)
	<-execCtx.Output().Done()
	time.Sleep(time.Millisecond * 100) // waiting for clearing execution in memory
	// load from storage
	result, ok = executor.GetResult(token)
	if !ok {
		t.Error("find result failed")
		return
	}
	if result.Finished != true || result.Result != "hello world" {
		t.Error("bad result")
	}
}

func TestTask_ExecutorReqId(t *testing.T) {
	var n int32 = 0
	cmd := NewTask(func(input *Input, progressFunc ProgressFunc) (interface{}, error) {
		atomic.AddInt32(&n, 1)
		return "hello " + input.Param().(string), nil
	}, "", TypeText)
	executor := NewExecutor(testStore)
	reqId := uuid.New().String()
	token, _ := executor.Execute(cmd, NewInput(context.Background(), "world").
		WithRequestTaskToken(reqId))
	if token.id != reqId {
		t.Error("bad token from reqId")
	}
	executor.WaitResult(token)
	if n != 1 {
		t.Error("should executed")
	}

	// duplicated reqId
	token, _ = executor.Execute(cmd, NewInput(context.Background(), "world").
		WithRequestTaskToken(reqId))
	if token.id != reqId {
		t.Error("bad token from reqId")
	}
	executor.WaitResult(token)
	if n != 1 {
		t.Error("should executed only once!")
	}

	// new executor, from store
	executor = NewExecutor(testStore)
	token, _ = executor.Execute(cmd, NewInput(context.Background(), "world").
		WithRequestTaskToken(reqId))
	if token.id != reqId {
		t.Error("bad token from reqId")
	}
	executor.WaitResult(token)
	if n != 1 {
		t.Error("should executed only once!")
	}
}

func TestTask_ExecutorMissing(t *testing.T) {
	executor := NewExecutor(testStore)
	_, ok := executor.GetResult(ExecutionToken{"1"})
	if ok {
		t.Error("should errors")
	}
	err := executor.Cancel(ExecutionToken{"1"})
	if err == nil {
		t.Error("should errors")
	}
}

func TestTask_Cancel(t *testing.T) {
	cmd := NewTask(func(input *Input, progressFunc ProgressFunc) (interface{}, error) {
		select {
		case <-input.Context().Done():
			return nil, input.Context().Err()
		case <-time.After(time.Second):
		}
		return "hello " + input.Param().(string), nil
	}, "", TypeText)

	canceled := make(chan struct{})
	cmd.OnCancel(func(execCtx *ExecutionContext) {
		fmt.Println("!!!!!!!!")
		close(canceled)
	})
	execCtx := NewInputExecutionContext(context.Background(), "world")
	_ = cmd.Execute(execCtx)
	execCtx.Cancel()
	_, err := execCtx.Output().Get()
	if err != context.Canceled {
		t.Error("should be canceled")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second * 5):
		t.Error("cmd cancel func not called")
	}
}

func TestTask_ExecutorCancel(t *testing.T) {
	executor := NewExecutor(testStore)
	cmd := NewTask(func(input *Input, progressFunc ProgressFunc) (interface{}, error) {
		select {
		case <-input.Context().Done():
			return nil, input.Context().Err()
		case <-time.After(time.Second):
		}
		return "hello " + input.Param().(string), nil
	}, "", TypeText)
	token, _ := executor.Execute(cmd, NewInput(context.Background(), "world"))
	_ = executor.Cancel(token)
	execCtx, ok := executor.getExecContext(token)
	if !ok {
		t.Error("execCtx not found")
		return
	}
	_, err := execCtx.Output().Get()
	if err != context.Canceled {
		t.Error("should be canceled")
	}
}

func TestTask_ExecutorTasks(t *testing.T) {
	executor := NewExecutor(testStore)
	cmd := NewTask(func(input *Input, progressFunc ProgressFunc) (interface{}, error) {
		select {
		case <-input.Context().Done():
			return nil, input.Context().Err()
		case <-time.After(time.Millisecond * 100):
		}
		return "hello " + input.Param().(string), nil
	}, "", TypeText)
	if len(executor.AllExecutions()) != 0 {
		t.Error("executor should be empty")
	}
	_, _ = executor.Execute(cmd, NewInput(context.Background(), "1"))
	_, _ = executor.Execute(cmd, NewInput(context.Background(), "2"))
	if len(executor.AllExecutions()) != 2 {
		t.Error("executor should have 2 executions")
	}
}

func TestTask_Wrap(t *testing.T) {
	task1 := WrapFunc(func(s string) (string, error) {
		return s + s, nil
	})
	ret, err := Execute(task1, "a")
	if err != nil || ret != "aa" {
		t.Error("failed")
	}

	task2 := WrapFunc(func() (string, error) {
		return "hello", nil
	})
	ret, err = Execute(task2, nil)
	if err != nil || ret != "hello" {
		t.Error("failed")
	}

	task3 := WrapFunc(func(ctx context.Context) (string, error) {
		return "hello", nil
	})
	ret, err = Execute(task3, nil)
	if err != nil || ret != "hello" {
		t.Error("failed")
	}

	task4 := WrapFunc(func(ctx context.Context, s string) string {
		return s
	})
	ret, err = Execute(task4, "test")
	if err != nil || ret != "test" {
		t.Error("failed")
	}

	task5 := WrapFunc(func(ctx context.Context) error {
		return errors.New("err")
	})
	ret, err = Execute(task5, nil)
	if err == nil || err.Error() != "err" || ret != nil {
		t.Error("failed")
	}
}
