// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package errors

import (
	"fmt"

	"golang.org/x/text/language"
)

// AgentError defines Agent specific errors.
// It implements error interface.
type AgentError struct {
	ErrorCode ErrorCode     // error code
	Args      []interface{} // args for error message formatting
}

func (e AgentError) Message(lang language.Tag) string {
	return GetMessage(lang, e.ErrorCode, e.Args)
}

func (e AgentError) DefaultMessage() string {
	return e.Message(defaultLanguage)
}

func (e AgentError) Error() string {
	return fmt.Sprintf("AgentError: code = %d, message = %s", e.ErrorCode.Code, e.DefaultMessage())
}

func Occur(errorCode ErrorCode, args ...interface{}) *AgentError {
	return &AgentError{
		ErrorCode: errorCode,
		Args:      args,
	}
}
