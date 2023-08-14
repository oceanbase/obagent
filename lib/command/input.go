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

// Annotation type alias for annotation data
// May contains some extra data out of input parameter like trace id, request id.
type Annotation map[string]interface{}

// Input command execution input.
// Used to pack context.Context, parameter and annotation into a single object passed to Command.
type Input struct {
	ctx        context.Context
	annotation Annotation
	param      interface{}
}

// NewInput Creates a new Input with context.Context and parameter
func NewInput(ctx context.Context, param interface{}) *Input {
	return &Input{ctx: ctx, annotation: make(Annotation), param: param}
}

// Context Returns the context.Context of an Input
func (input *Input) Context() context.Context {
	return input.ctx
}

// Annotation Returns the Annotation of an Input
func (input *Input) Annotation() Annotation {
	return input.annotation
}

// WithAnnotation Adds an Annotation key-value into an Input
func (input *Input) WithAnnotation(key string, value interface{}) *Input {
	input.annotation[key] = value
	return input
}

// Param Returns the parameter of an Input
func (input *Input) Param() interface{} {
	return input.param
}

func (input *Input) WithRequestTaskToken(requestId string) *Input {
	return input.WithAnnotation(RequestTaskTokenKey, requestId)
}
