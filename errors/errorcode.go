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

import "net/http"

type ErrorKind = int

const (
	badRequest      ErrorKind = http.StatusBadRequest
	illegalArgument ErrorKind = http.StatusBadRequest
	notFound        ErrorKind = http.StatusNotFound
	unexpected      ErrorKind = http.StatusInternalServerError
	notImplemented  ErrorKind = http.StatusNotImplemented
)

type ErrorCode struct {
	Code int
	Kind ErrorKind
	key  string
}

var errorCodes []ErrorCode

func NewErrorCode(code int, kind ErrorKind, key string) ErrorCode {
	errorCode := ErrorCode{
		Code: code,
		Kind: kind,
		key:  key,
	}
	errorCodes = append(errorCodes, errorCode)
	return errorCode
}

var (
	ErrBadRequest      = NewErrorCode(1000, badRequest, "err.bad.request")
	ErrIllegalArgument = NewErrorCode(1001, illegalArgument, "err.illegal.argument")
	ErrUnexpected      = NewErrorCode(1002, unexpected, "err.unexpected")
)
