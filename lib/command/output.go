/*
 * Copyright (c) 2023 OceanBase
 * OBAgent is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package command

import (
	"sync"
	"sync/atomic"
	"time"
)

// Output Command execution output, used to handle command result, error, notify command finish event...
type Output struct {
	lock       sync.Mutex
	result     interface{}
	err        error
	ok         bool
	finishChan chan struct{}
	finished   bool
	progress   atomic.Value
}

// OutputStatus A snapshot of execution Output
type OutputStatus struct {
	//Finished The execution is finished or not
	Finished bool
	//Ok The execution is finished successfully or not
	Ok bool
	//Result Success result
	Result interface{}
	//Err Fail error
	Err string
	//Progress progress data
	Progress interface{}
}

func newOutput() *Output {
	return &Output{
		result:     nil,
		err:        nil,
		finishChan: make(chan struct{}),
		finished:   false,
	}
}

// FinishOk mark the command execution succeed with a result
// Returns AlreadyFinishedErr when execution was already finished
func (output *Output) FinishOk(result interface{}) error {
	output.lock.Lock()
	defer output.lock.Unlock()
	if output.finished {
		return AlreadyFinishedErr
	}
	output.finished = true
	output.ok = true
	output.result = result
	close(output.finishChan)
	return nil
}

// FinishErr mark the command execution failed with an error
// Returns AlreadyFinishedErr when execution was already finished
func (output *Output) FinishErr(err error) error {
	output.lock.Lock()
	defer output.lock.Unlock()
	if output.finished {
		return AlreadyFinishedErr
	}
	output.finished = true
	output.ok = false
	output.err = err
	close(output.finishChan)
	return nil
}

// Finished return the command execution is finished or not
func (output *Output) Finished() bool {
	output.lock.Lock()
	ret := output.finished
	output.lock.Unlock()
	return ret
}

// UpdateProgress Update command execution progress.
func (output *Output) UpdateProgress(progress interface{}) {
	output.progress.Store(progress)
}

// Progress Get current command execution progress.
func (output *Output) Progress() interface{} {
	return output.progress.Load()
}

// Get Get command execution result or error synchronously.
func (output *Output) Get() (interface{}, error) {
	<-output.finishChan

	output.lock.Lock()
	defer output.lock.Unlock()
	return output.result, output.err
}

// GetWithTimeout Get command execution result or error synchronously with a timeout.
// Will return a timeout error when the execution not return a result after the timeout duration.
func (output *Output) GetWithTimeout(timeout time.Duration) (interface{}, error) {
	ch := output.finishChan
	select {
	case <-ch:
		return output.result, output.err
	case <-time.After(timeout):
		return nil, TimeoutErr
	}
}

// Ok Return whether the command execution is finished and with a success result
func (output *Output) Ok() bool {
	output.lock.Lock()
	ret := output.ok
	output.lock.Unlock()
	return ret
}

// Done Return a chan to wait the command execution finished.
// When execution finished, the chan will be closed.
func (output *Output) Done() <-chan struct{} {
	return output.finishChan
}

// Status Return a snapshot of the execution Output. See OutputStatus
func (output *Output) Status() OutputStatus {
	output.lock.Lock()
	errStr := ""
	if output.err != nil {
		errStr = output.err.Error()
	}
	ret := OutputStatus{
		Finished: output.finished,
		Ok:       output.ok,
		Result:   output.result,
		Err:      errStr,
		Progress: output.Progress(),
	}
	output.lock.Unlock()
	return ret
}
