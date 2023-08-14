/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
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

// Command Abstract command interface,
// represents an executable object which accept a parameter and return a result or an error.
// Execution result may be get asynchronously.
type Command interface {
	//Execute Execute the command with ExecutionContext
	Execute(*ExecutionContext) error

	//DefaultParam returns the default parameter of the command.
	//default parameter may be used to deserialize input parameter by a command handler from HTTP request.
	DefaultParam() interface{}

	//ResponseType Returns response DataType of the command
	ResponseType() DataType
}

// DataType an enum type of data type of command result
type DataType string

const (
	//TypeStructured Structured result, may be struct, map, etc.
	//Structured data Will be serialize when storing result or deserialize when loading result.
	TypeStructured DataType = "structured"

	//TypeText plain text result, may be a string.
	//Text data won't be serialize or deserialize when storing or loading result.
	TypeText DataType = "text"

	//TypeBinary binary result, may be a []byte.
	//Binary data won't be serialize or deserialize when storing or loading result.
	TypeBinary DataType = "binary"
)

func Execute(command Command, param interface{}) (interface{}, error) {
	execCtx := NewInputExecutionContext(context.Background(), param)
	err := command.Execute(execCtx)
	if err != nil {
		return nil, err
	}
	return execCtx.Output().Get()
}
