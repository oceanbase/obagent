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

package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/mgragent"
	"github.com/oceanbase/obagent/errors"
	http2 "github.com/oceanbase/obagent/lib/http"
)

func Test_RouteHandler(t *testing.T) {
	type args struct {
		url string
	}
	type want struct {
		successful bool
		statusCode int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "example1",
			args: args{
				url: "http://127.0.0.1:62888/api/example/1",
			},
			want: want{
				successful: true,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "example2",
			args: args{
				url: "http://127.0.0.1:62888/api/example/2",
			},
			want: want{
				successful: true,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "example3",
			args: args{
				url: "http://127.0.0.1:62888/api/example/3",
			},
			want: want{
				successful: true,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "example4",
			args: args{
				url: "http://127.0.0.1:62888/api/example/4",
			},
			want: want{
				successful: false,
				statusCode: http.StatusBadRequest,
			},
		},
	}
	server := NewServer(config.AgentVersion, mgragent.ServerConfig{})
	InitExampleRoutes(server.Router)
	handler := func(w http.ResponseWriter, r *http.Request) {
		server.Router.ServeHTTP(w, r)
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			req := httptest.NewRequest("GET", tt.args.url, nil)
			w := httptest.NewRecorder()
			handler(w, req)

			resp := w.Result()
			var b bytes.Buffer
			io.Copy(&b, resp.Body)

			var successResponse http2.OcpAgentResponse
			_ = json.Unmarshal(b.Bytes(), &successResponse)

			So(resp.StatusCode, ShouldEqual, tt.want.statusCode)
			So(successResponse.Successful, ShouldEqual, tt.want.successful)
			So(successResponse.Status, ShouldEqual, tt.want.statusCode)
		})
	}
}

// only for test
func InitExampleRoutes(r *gin.Engine) {
	v1 := r.Group("/api/example")

	v1.GET("/1", exampleHandler1)
	v1.GET("/2", exampleHandler2)
	v1.GET("/3", exampleHandler3)
	v1.GET("/4", exampleHandler4)
}

var exampleHandler1 = func(c *gin.Context) {
	data, err := singleExample()
	sendResponse(c, data, err)
}

var exampleHandler2 = func(c *gin.Context) {
	data, err := iterableExample()
	sendResponse(c, data, err)
}

var exampleHandler3 = func(c *gin.Context) {
	err := noDataExample()
	sendResponse(c, nil, err)
}

var exampleHandler4 = func(c *gin.Context) {
	data, err := errorExample()
	sendResponse(c, data, err)
}

func singleExample() (string, *errors.OcpAgentError) {
	return "this is data", nil
}

func iterableExample() ([]string, *errors.OcpAgentError) {
	return []string{"data1", "data2", "data3"}, nil
}

func noDataExample() *errors.OcpAgentError {
	return nil
}

func errorExample() (string, *errors.OcpAgentError) {
	return "", errors.Occur(errors.ErrBadRequest)
}

func sendResponse(c *gin.Context, data interface{}, err error) {
	resp := http2.BuildResponse(data, err)
	c.Set(common.OcpAgentResponseKey, resp)
}
