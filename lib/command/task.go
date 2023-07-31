package command

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ProgressFunc function to update execution progress. See Output.UpdateProgress
type ProgressFunc func(interface{})

// TaskFunc A function represents a Task. To simplify Command
type TaskFunc func(*Input, ProgressFunc) (interface{}, error)

// CancelFunc function to be executed when cmd canceled
type CancelFunc func(*ExecutionContext)

const RequestTaskTokenKey = "reqTaskToken"

var taskIdPattern = regexp.MustCompile("^[0-9a-zA-Z_-]{8,64}$")

// Task a Command implement that turns synchronize function call to a Command.
type Task struct {
	fn           TaskFunc
	cancel       CancelFunc
	defaultParam interface{}
	dataType     DataType
}

// NewTask create a new Task
func NewTask(fn TaskFunc, defaultParam interface{}, dataType DataType) *Task {
	return &Task{fn: fn, defaultParam: defaultParam, dataType: dataType}
}

func (t *Task) DefaultParam() interface{} {
	return t.defaultParam
}

func (t *Task) Execute(execContext *ExecutionContext) error {
	go func() {
		defer func() {
			e := recover()
			if e != nil {
				var err error
				var ok bool
				if err, ok = e.(error); ok {
					_ = execContext.Output().FinishErr(err)
				} else {
					err = fmt.Errorf("panic: %+v. task: %+v, input: %+v", e, t, execContext.input)
					_ = execContext.Output().FinishErr(err)
				}
			}
		}()

		result, err := t.fn(execContext.Input(), execContext.Output().UpdateProgress)
		if err == nil {
			_ = execContext.Output().FinishOk(result)
		} else {
			_ = execContext.Output().FinishErr(err)
		}
	}()
	go func() {
		<-execContext.Output().Done()
		if execContext.Output().err == context.Canceled {
			<-execContext.input.ctx.Done()
			if execContext.input.ctx.Err() == context.Canceled && t.cancel != nil {
				t.cancel(execContext)
			}
		}
	}()
	return nil
}

func (t *Task) ResponseType() DataType {
	return t.dataType
}

// OnCancel sets a callback that will be called when execution canceled.
func (t *Task) OnCancel(cancel CancelFunc) {
	t.cancel = cancel
}

// Execution a running command execution.
// A command can be executed many times. Execution will only lives while a command executing,
// and will be destroyed after execution finished.
type Execution struct {
	cmd     Command
	ctx     *ExecutionContext
	token   ExecutionToken
	startAt time.Time
	endAt   time.Time
}

// Command returns the cmd of the execution
func (e Execution) Command() Command {
	return e.cmd
}

// ExecutionContext returns the cmd of the execution
func (e Execution) ExecutionContext() *ExecutionContext {
	return e.ctx
}

// Executor runs Command s and maintains Command s' status
// Executor can run Command background and return a ExecutionToken
type Executor struct {
	executions sync.Map
	id         int64
	nowFunc    func() int64
	store      StatusStore
}

// NewExecutor create a new Executor with a StatusStore
func NewExecutor(store StatusStore) *Executor {
	ret := &Executor{
		nowFunc: func() int64 {
			return time.Now().UnixNano()
		},
		store: store,
	}
	ret.id = ret.nowFunc()
	return ret
}

// ExecutionToken A token returned to caller of Executor.Execute.
// Caller can use it lately to query execution result.
type ExecutionToken struct {
	id string
}

// ExecutionTokenFromString convert a string to a ExecutionToken
func ExecutionTokenFromString(s string) ExecutionToken {
	return ExecutionToken{
		id: s,
	}
}

// String convert a ExecutionToken to a string
func (t ExecutionToken) String() string {
	return t.id
}

func GenerateTaskId() string {
	return uuid.New().String()
}

// Execute Executes A Command with an Input.
// Returns an ExecutionToken A token returned to caller.
// Caller can use it lately to query execution result.
// When a request id provided via input Annotation "taskReqId",
// Executor will use the request id as the ExecutionToken.
// Execute call with same request id will be executed only once and return the previous execution's ExecutionToken.
// That means, when using request id, the request id MUST be unique all over the time.
// Execution result will be stored persistently and can be queried after process restart.
func (ex *Executor) Execute(cmd Command, input *Input) (ExecutionToken, error) {
	var id string
	id, fromReq := ex.getTaskId(input)

	if fromReq {
		// check request id exists or not
		// return prev token when exists
		token := ExecutionTokenFromString(id)
		if _, ok := ex.GetResult(token); ok {
			return token, nil
		}
	}
	execContext := NewExecutionContext(input)
	ret := ExecutionToken{id: id}
	execution := &Execution{cmd: cmd, ctx: execContext, token: ret}
	err := ex.store.Create(ret, execution)
	if err != nil {
		if err == ExecutionAlreadyExistsErr {
			return ret, nil
		}
		return ExecutionToken{}, err
	}
	err = cmd.Execute(execContext)
	if err != nil {
		return ExecutionToken{}, err
	}
	execution.startAt = time.Now()
	ex.executions.Store(id, execution)

	err = ex.store.Store(ret, execution)
	if err != nil {
		return ExecutionToken{}, err
	}
	go func(token ExecutionToken, execution *Execution) {
		<-execution.ExecutionContext().Output().Done()
		execution.endAt = time.Now()
		err1 := ex.store.Store(token, execution)
		if err1 != nil {
			//
		}
		ex.executions.Delete(id)
	}(ret, execution)
	return ret, nil
}

func (ex *Executor) Detach(token ExecutionToken) {
	ex.executions.Delete(token.id)
}

// getTaskId get request id from input if exists, or generates a new token locally.
func (ex *Executor) getTaskId(input *Input) (string, bool) {
	if reqIdVar, ok := input.Annotation()[RequestTaskTokenKey]; ok {
		if reqId, ok := reqIdVar.(string); ok {
			if taskIdPattern.MatchString(reqId) {
				return reqId, true
			}
		}
	}
	return GenerateTaskId(), false
}

// GetResult Gets execution result. It will return immediately with the OutputStatus and a
// bool means the execution specified by the token exists or not.
func (ex *Executor) GetResult(token ExecutionToken) (OutputStatus, bool) {
	execCtx, ok := ex.getExecContext(token)
	if ok {
		return execCtx.Output().Status(), true
	}
	status, err := ex.store.Load(token)
	if err != nil {
		return OutputStatus{}, false
	}
	return OutputStatus{
		Finished: status.Finished,
		Ok:       status.Ok,
		Result:   status.Result,
		Err:      status.Err,
		Progress: status.Progress,
	}, true
}

// Cancel cancels the execution specified by the token
// If execution has finished or not exists, will return ExecutionNotFoundErr
func (ex *Executor) Cancel(token ExecutionToken) error {
	h, ok := ex.GetExecution(token)
	if !ok {
		return ExecutionNotFoundErr
	}
	h.ctx.Cancel()
	return nil
}

// GetExecution Gets the execution object. It will return immediately with the Execution and a
// bool means the execution is still running or not.
func (ex *Executor) GetExecution(token ExecutionToken) (*Execution, bool) {
	loaded, ok := ex.executions.Load(token.id)
	if ok {
		h := loaded.(*Execution)
		return h, true
	}
	return nil, false
}

// WaitResult Wait for execution result synchronously. Returns the
// OutputStatus after execution finished when execution exists.
// If the execution not exists, it will returns with a false second return value immediately.
func (ex *Executor) WaitResult(token ExecutionToken) (OutputStatus, bool) {
	if e, ok := ex.GetExecution(token); ok {
		_, _ = e.ExecutionContext().Output().Get()
	}
	return ex.GetResult(token)
}

func (ex *Executor) getExecContext(token ExecutionToken) (*ExecutionContext, bool) {
	h, ok := ex.GetExecution(token)
	if !ok {
		return nil, false
	}
	return h.ctx, true
}

// AllExecutions returns all running executions
func (ex *Executor) AllExecutions() []*Execution {
	var ret []*Execution
	ex.executions.Range(func(k interface{}, v interface{}) bool {
		ret = append(ret, v.(*Execution))
		return true
	})
	return ret
}

// WrapFunc converts a function into a Task
// form of function can be one of:
//
// func() RetType
// func() (RetType, error)
// func() error
// func(arg SomeType) RetType
// func(arg SomeType) (RetType, error)
// func(arg SomeType) error
// func(ctx context.Context, arg SomeType) RetType
// func(ctx context.Context, arg SomeType) (RetType, error)
// func(ctx context.Context, arg SomeType) error
// func(ctx context.Context) RetType
// func(ctx context.Context) (RetType, error)
// func(ctx context.Context) error
func WrapFunc(fn interface{}) *Task {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()
	if fnType.Kind() != reflect.Func {
		panic("not a valid function")
	}
	if fnType.NumIn() > 2 && fnType.NumOut() > 2 {
		panic("not a valid function")
	}
	var tCtx = context.Background()
	var tErr = errors.New("")
	errType := reflect.TypeOf(&tErr).Elem()
	ctxType := reflect.TypeOf(&tCtx).Elem()

	ctxOffset := -1
	argOffset := -1
	retOffset := -1
	errOffset := -1
	var defaultParam interface{} = nil
	if fnType.NumIn() == 1 {
		if fnType.In(0).Implements(ctxType) {
			ctxOffset = 0
		} else {
			argOffset = 0
		}
	} else if fnType.NumIn() == 2 {
		ctxOffset = 0
		argOffset = 1
	}
	if argOffset >= 0 {
		defaultParam = reflect.New(fnType.In(argOffset)).Elem().Interface()
	}
	if fnType.NumOut() == 1 {
		out0 := fnType.Out(0)
		if out0.Implements(errType) {
			errOffset = 0
		} else {
			retOffset = 0
		}
	} else if fnType.NumOut() == 2 {
		retOffset = 0
		errOffset = 1
	}

	return &Task{fn: func(input *Input, progressFunc ProgressFunc) (interface{}, error) {
		//recover()
		var retValues []reflect.Value
		inValues := make([]reflect.Value, fnType.NumIn())
		if ctxOffset >= 0 {
			inValues[ctxOffset] = reflect.ValueOf(input.Context())
		}
		if argOffset >= 0 {
			inValues[argOffset] = reflect.ValueOf(input.Param())
		}
		retValues = fnValue.Call(inValues)
		if retValues == nil {
			return nil, BadTaskFunc
		}
		var ret interface{} = nil
		var err error = nil
		if retOffset >= 0 {
			ret = retValues[retOffset].Interface()
		}
		if errOffset >= 0 {
			errValue := retValues[errOffset]
			reflect.ValueOf(&err).Elem().Set(errValue)
		}
		return ret, err
	}, defaultParam: defaultParam, dataType: TypeStructured}
}
