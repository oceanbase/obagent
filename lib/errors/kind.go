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

package errors

import "fmt"

// Kind A wide catalog of error code. A Kind implies general error cause and handling suggestion.
type Kind struct {
	Id       int
	Name     string
	HttpCode int
}

// Code A concrete type of error. Such as "file not exists"
type Code struct {
	Kind            Kind
	Module          string
	Name            string
	MessageTemplate string
}

var (
	Ok = Kind{Id: 0, Name: "OK", HttpCode: 200}

	Cancelled          = Kind{Id: 50, Name: "CANCELLED", HttpCode: 499}
	InvalidArgument    = Kind{Id: 51, Name: "INVALID_ARGUMENT", HttpCode: 400}
	NotFound           = Kind{Id: 52, Name: "NOT_FOUND", HttpCode: 404}
	AlreadyExists      = Kind{Id: 53, Name: "ALREADY_EXISTS", HttpCode: 409}
	PermissionDenied   = Kind{Id: 54, Name: "PERMISSION_DENIED", HttpCode: 403}
	Unauthenticated    = Kind{Id: 55, Name: "UNAUTHENTICATED", HttpCode: 401}
	ResourceExhausted  = Kind{Id: 56, Name: "RESOURCE_EXHAUSTED", HttpCode: 429}
	FailedPrecondition = Kind{Id: 57, Name: "FAILED_PRECONDITION", HttpCode: 400}
	OutOfRange         = Kind{Id: 58, Name: "OUT_OF_RANGE", HttpCode: 400}
	Aborted            = Kind{Id: 59, Name: "ABORTED", HttpCode: 409}

	Unknown          = Kind{Id: 10, Name: "UNKNOWN", HttpCode: 500}
	DeadlineExceeded = Kind{Id: 11, Name: "DEADLINE_EXCEEDED", HttpCode: 504}
	Internal         = Kind{Id: 12, Name: "INTERNAL", HttpCode: 500}
	Unavailable      = Kind{Id: 13, Name: "UNAVAILABLE", HttpCode: 503}
	DataLoss         = Kind{Id: 14, Name: "DATA_LOSS", HttpCode: 500}

	byName = map[string]Kind{
		Ok.Name:                 Ok,
		Cancelled.Name:          Cancelled,
		InvalidArgument.Name:    InvalidArgument,
		NotFound.Name:           NotFound,
		AlreadyExists.Name:      AlreadyExists,
		PermissionDenied.Name:   PermissionDenied,
		Unauthenticated.Name:    Unauthenticated,
		ResourceExhausted.Name:  ResourceExhausted,
		FailedPrecondition.Name: FailedPrecondition,
		OutOfRange.Name:         OutOfRange,
		Aborted.Name:            Aborted,
		Unknown.Name:            Unknown,
		DeadlineExceeded.Name:   DeadlineExceeded,
		Internal.Name:           Internal,
		Unavailable.Name:        Unavailable,
		DataLoss.Name:           DataLoss,
	}
)

func KindByName(name string) Kind {
	return byName[name]
}

func (k Kind) NewCode(module, code string) Code {
	return Code{
		Kind:            k,
		Module:          module,
		Name:            code,
		MessageTemplate: "",
	}
}

func (c Code) NewError(values ...interface{}) *Error {
	return &Error{
		code:   c,
		values: values,
		cause:  nil,
		stack:  "",
	}
}

func (c Code) Equals(o Code) bool {
	return c.Module == o.Module && c.Name == o.Name && c.Kind.Id == o.Kind.Id
}

func (c Code) Format(values ...interface{}) string {
	if len(values) == 0 {
		return c.MessageTemplate
	}
	if c.MessageTemplate == "" {
		return fmt.Sprintf("%+v", values)
	}
	return fmt.Sprintf(c.MessageTemplate, values...)
}

func (c Code) WithMessageTemplate(tpl string) Code {
	c.MessageTemplate = tpl
	return c
}
