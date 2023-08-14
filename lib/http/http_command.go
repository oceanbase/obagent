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

package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	errors2 "github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/command"
	"github.com/oceanbase/obagent/lib/errors"
	"github.com/oceanbase/obagent/lib/system"
	"github.com/oceanbase/obagent/lib/trace"
)

type TaskToken struct {
	Token string `json:"token"`
}

var serverIp, _ = system.SystemImpl{}.GetLocalIpAddress()

type reqKey string

const (
	RequestTimeKey reqKey = "request_time"
)

func NewFuncHandler(fun interface{}) http.HandlerFunc {
	return NewHandler(command.WrapFunc(fun))
}

func NewHandler(cmd command.Command) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		ctx := context.WithValue(trace.ContextWithTraceId(request), RequestTimeKey, time.Now())
		request = request.WithContext(ctx)

		cmdParam, err := parseJsonParam(request.Body, cmd.DefaultParam())
		if err != nil {
			writeError(err, request, writer)
			return
		}
		log.WithContext(ctx).Infof("handling command request %s %v", request.URL.Path, cmdParam)
		execContext := command.NewInputExecutionContext(ctx, cmdParam)
		err = cmd.Execute(execContext)
		if err != nil {
			writeError(err, request, writer)
			return
		}
		result, err := execContext.Output().Get()
		if err != nil {
			writeError(err, request, writer)
			return
		}
		writeOk(result, request, writer)
	}
}

func parseJsonParam(reader io.Reader, defaultParam interface{}) (interface{}, error) {
	v := reflect.ValueOf(defaultParam)
	if !v.IsValid() {
		return nil, nil
	}
	v1 := reflect.New(v.Type())
	v1.Elem().Set(v)
	err := json.NewDecoder(reader).Decode(v1.Interface())
	if err != nil {
		return nil, err
	}
	return v1.Elem().Interface(), nil
}

func writeData(ctx context.Context, result interface{}, writer http.ResponseWriter) {
	data, _ := json.Marshal(result)
	writer.Header().Set("Connection", "close")
	_, err := writer.Write(data)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("write data to client got error. maybe client closed connection.")
	}
}

func writeOk(result interface{}, request *http.Request, writer http.ResponseWriter) {
	envelop := wrapEnvelop(result, nil, request)
	ctx := trace.ContextWithTraceId(request)
	writeData(ctx, envelop, writer)
	log.WithContext(ctx).Infof("command request %s succeed", request.URL.Path)
}

func wrapEnvelop(result interface{}, err error, request *http.Request) OcpAgentResponse {
	value := request.Context().Value(RequestTimeKey)
	var requestTime time.Time
	var ok bool
	var duration int
	if requestTime, ok = value.(time.Time); ok {
		duration = int(time.Since(requestTime).Milliseconds())
	}
	var apiErr *ApiError
	success := true
	code := 0
	statusCode := 200
	if err != nil {
		success = false
		statusCode = errors.HttpCode(err, http.StatusInternalServerError)
		if e, ok := err.(*errors2.OcpAgentError); ok {
			code = e.ErrorCode.Code
			statusCode = e.ErrorCode.Kind
		} else if e, ok := err.(*errors.Error); ok {
			code = e.Kind().HttpCode
			statusCode = code
		}
		apiErr = &ApiError{
			Code:    code,
			Message: err.Error(),
		}
	}
	return OcpAgentResponse{
		Successful: success,
		Timestamp:  requestTime,
		Duration:   duration,
		TraceId:    trace.GetTraceId(request),
		Server:     serverIp,
		Data:       result,
		Error:      apiErr,
		Status:     statusCode,
	}
}

func writeError(err error, request *http.Request, writer http.ResponseWriter) {
	envelop := wrapEnvelop(nil, err, request)
	writer.WriteHeader(envelop.Status)
	ctx := trace.ContextWithTraceId(request)
	writeData(ctx, envelop, writer)
	log.WithContext(ctx).Warnf("command request %s got error: %v", request.URL.Path, err)
}
