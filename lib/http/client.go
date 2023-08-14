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
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

func CanConnect(network, addr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout(network, addr, timeout)
	ret := err == nil
	if conn != nil {
		_ = conn.Close()
	}
	return ret
}

type ApiClient struct {
	protocol string
	host     string
	hc       *http.Client
}

func NewSocketApiClient(socket string, timeout time.Duration) *ApiClient {
	return &ApiClient{
		protocol: "http",
		host:     "socket",
		hc: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DisableKeepAlives: true,
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socket)
				},
			},
		},
	}
}

const ContentTypeJson = "application/json"

func (ac *ApiClient) Call(api string, param interface{}, retPtr interface{}) error {
	var inputData []byte
	var err error
	if param != nil {
		inputData, err = json.Marshal(param)
		if err != nil {
			return EncodeRequestFailedErr.NewError().WithCause(err)
		}
	}
	reader := bytes.NewReader(inputData)
	resp, err := ac.hc.Post(ac.url(api), ContentTypeJson, reader)
	if err != nil {
		return ApiRequestFailedErr.NewError(api).WithCause(err)
	}
	defer ac.hc.CloseIdleConnections()
	defer resp.Body.Close()

	envelop := OcpAgentResponse{
		Data: retPtr,
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ApiRequestFailedErr.NewError(api).WithCause(err)
	}
	cpResp := make([]byte, len(bs))
	copy(cpResp, bs)
	respReader := bytes.NewReader(bs)

	if resp.StatusCode != http.StatusOK {
		err = json.NewDecoder(respReader).Decode(&envelop)
		if err == nil {
			return ApiRequestGotFailResultErr.NewError(api, envelop.Error.Code, envelop.Error.Message)
		}
		return ApiRequestGotFailResultErr.WithMessageTemplate("api %s, resp code %d, status %s, decode resp %s err %s").
			NewError(api, resp.StatusCode, resp.Status, cpResp, err)
	}
	err = json.NewDecoder(respReader).Decode(&envelop)
	if err != nil {
		return DecodeResultFailedErr.WithMessageTemplate("api %s, decode resp %s").NewError(api, cpResp).WithCause(err)
	}
	if envelop.Successful {
		return nil
	}
	return ApiRequestGotFailResultErr.NewError(api, envelop.Error.Code, envelop.Error.Message)
}

func (ac *ApiClient) url(api string) string {
	url := ac.protocol + "://" + ac.host
	if !strings.HasPrefix(api, "/") {
		url += "/"
	}
	return url + api
}
