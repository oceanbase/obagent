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

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/oceanbase/obagent/errors"
)

// OcpAgentResponse defines basic API return structure for HTTP responses.
type OcpAgentResponse struct {
	Successful bool        `json:"successful"`      // Whether request successful or not
	Timestamp  time.Time   `json:"timestamp"`       // Request handling timestamp (server time)
	Duration   int         `json:"duration"`        // Request handling time cost (ms)
	Status     int         `json:"status"`          // HTTP status code
	TraceId    string      `json:"traceId"`         // Request trace ID, contained in server logs
	Server     string      `json:"server"`          // Server's internal IP address
	Data       interface{} `json:"data,omitempty"`  // Data payload when response is successful
	Error      *ApiError   `json:"error,omitempty"` // Error payload when response is failed
}

type ocpAgentResponseJson struct {
	Successful bool            `json:"successful"`      // Whether request successful or not
	Timestamp  time.Time       `json:"timestamp"`       // Request handling timestamp (server time)
	Duration   int             `json:"duration"`        // Request handling time cost (ms)
	Status     int             `json:"status"`          // HTTP status code
	TraceId    string          `json:"traceId"`         // Request trace ID, contained in server logs
	Server     string          `json:"server"`          // Server's internal IP address
	Data       json.RawMessage `json:"data,omitempty"`  // Data payload when response is successful
	Error      *ApiError       `json:"error,omitempty"` // Error payload when response is failed
}

func (r *OcpAgentResponse) UnmarshalJSON(b []byte) error {
	j := ocpAgentResponseJson{}
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	r.Successful = j.Successful
	r.Timestamp = j.Timestamp
	r.Duration = j.Duration
	r.Status = j.Status
	r.TraceId = j.TraceId
	r.Server = j.Server
	r.Error = j.Error
	v := reflect.ValueOf(r.Data)
	if !v.IsValid() {
		err = json.Unmarshal(j.Data, &r.Data)
		if err != nil {
			return err
		}
	} else if v.Type().Kind() == reflect.Ptr {
		err = json.Unmarshal(j.Data, r.Data)
		if err != nil {
			return err
		}
	} else {
		tmp := reflect.New(v.Type()).Interface()
		err = json.Unmarshal(j.Data, tmp)
		if err != nil {
			return err
		}
		r.Data = reflect.ValueOf(tmp).Elem().Interface()
	}
	return nil
}

type IterableData struct {
	Contents interface{} `json:"contents"`
}

type ApiError struct {
	Code      int           `json:"code"`                // Error code
	Message   string        `json:"message"`             // Error message
	SubErrors []interface{} `json:"subErrors,omitempty"` // Sub errors
}

func (a ApiError) String() string {
	if len(a.SubErrors) == 0 {
		return fmt.Sprintf("{Code:%v, Message:%v}", a.Code, a.Message)
	} else {
		return fmt.Sprintf("{Code:%v, Message:%v, SubErrors:%+v}", a.Code, a.Message, a.SubErrors)
	}
}

type ApiFieldError struct {
	Tag     string `json:"tag"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewApiFieldError(fieldError validator.FieldError) ApiFieldError {
	return ApiFieldError{
		Tag:     fieldError.Tag(),
		Field:   fieldError.Field(),
		Message: fieldError.Error(),
	}
}

type ApiUnknownError struct {
	Error error `json:"error"`
}

func NewSuccessResponse(data interface{}) OcpAgentResponse {
	return OcpAgentResponse{
		Successful: true,
		Timestamp:  time.Now(),
		Status:     http.StatusOK,
		Data:       data,
		Error:      nil,
	}
}

func NewErrorResponse(err *errors.OcpAgentError) OcpAgentResponse {
	return OcpAgentResponse{
		Successful: false,
		Timestamp:  time.Now(),
		Status:     err.ErrorCode.Kind,
		Data:       nil,
		Error: &ApiError{
			Code:    err.ErrorCode.Code,
			Message: err.DefaultMessage(),
		},
	}
}

func NewSubErrorsResponse(subErrors []interface{}) OcpAgentResponse {
	allValidationError := true
	for _, subError := range subErrors {
		if _, ok := subError.(ApiFieldError); !ok {
			allValidationError = false
		}
	}

	var status int
	var code int
	var message string
	if allValidationError {
		status = errors.ErrBadRequest.Kind
		code = errors.ErrBadRequest.Code
		message = "validation error"
	} else {
		status = errors.ErrUnexpected.Kind
		code = errors.ErrUnexpected.Code
		message = "unhandled error"
	}

	return OcpAgentResponse{
		Successful: false,
		Timestamp:  time.Now(),
		Status:     status,
		Data:       nil,
		Error: &ApiError{
			Code:      code,
			Message:   message,
			SubErrors: subErrors,
		},
	}
}

func BuildResponse(data interface{}, err error) OcpAgentResponse {
	agenterr, ok := err.(*errors.OcpAgentError)
	if !ok && err != nil {
		agenterr = errors.Occur(errors.ErrUnexpected, err)
	}
	if agenterr != nil {
		return NewErrorResponse(agenterr)
	}

	if data != nil && reflect.TypeOf(data).Kind() == reflect.Slice {
		iterableData := IterableData{Contents: data}
		return NewSuccessResponse(iterableData)
	} else {
		return NewSuccessResponse(data)
	}
}
