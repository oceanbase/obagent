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

import "context"

// ExecutionContext Execution Context for a Command to be executed with.
// Contains Input, Output, and can be canceled by a context.CancelFunc
type ExecutionContext struct {
	input  *Input
	output *Output
	cancel context.CancelFunc
	// logger
}

// NewExecutionContext Creates a new ExecutionContext with an Input
func NewExecutionContext(input *Input) *ExecutionContext {
	ret := &ExecutionContext{
		input:  input,
		output: newOutput(),
		cancel: nil,
	}
	ret.input.ctx, ret.cancel = context.WithCancel(ret.input.ctx)
	return ret
}

// NewInputExecutionContext Create a new ExecutionContext with context.Context and a param
// A shortcut function of NewExecutionContext(NewInput(ctx, param))
func NewInputExecutionContext(ctx context.Context, param interface{}) *ExecutionContext {
	return NewExecutionContext(NewInput(ctx, param))
}

// Input Returns Input of the ExecutionContext
func (execCtx *ExecutionContext) Input() *Input {
	return execCtx.input
}

// Output Returns Output of the ExecutionContext
func (execCtx *ExecutionContext) Output() *Output {
	return execCtx.output
}

// Cancel Cancel execution of the ExecutionContext.
// Caller will get a context.Canceled error when getting result.
func (execCtx *ExecutionContext) Cancel() {
	err := execCtx.Output().FinishErr(context.Canceled) // use custom canceled error?
	if err == nil {
		execCtx.cancel()
	}
}
