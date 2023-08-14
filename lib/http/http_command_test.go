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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type St struct {
	A string
	B int
}

func Test_parseJsonParam(t *testing.T) {
	var str string
	v, err := parseJsonParam(strings.NewReader(`"str"`), str)
	if err != nil {
		t.Error(err)
	}
	if v != "str" {
		t.Error("bad value")
	}

	v, err = parseJsonParam(strings.NewReader(`"str"`), &str)
	if err != nil {
		t.Error(err)
	}
	if *(v.(*string)) != "str" {
		t.Error("bad value")
	}

	v, err = parseJsonParam(strings.NewReader(`null`), &str)
	if err != nil {
		t.Error(err)
	}
	if v.(*string) != nil {
		println(*v.(*string))
		t.Error("bad value")
	}

	v, err = parseJsonParam(strings.NewReader(`"str2"`), &str)
	if err != nil {
		t.Error(err)
	}
	if *v.(*string) != "str2" {
		t.Error("bad value")
	}

	var sl []string
	var m map[string]string

	v, err = parseJsonParam(strings.NewReader(`["a"]`), sl)
	if err != nil {
		t.Error(err)
	}
	if len(v.([]string)) != 1 {
		t.Error("bad value")
	}

	v, err = parseJsonParam(strings.NewReader(`{"a":"1"}`), m)
	if err != nil {
		t.Error(err)
	}
	if len(v.(map[string]string)) != 1 {
		t.Error("bad value")
	}

	var st = St{
		A: "A",
		B: 123,
	}
	v, err = parseJsonParam(strings.NewReader(`{"B":234}`), st)
	if err != nil {
		t.Error(err)
	}
	if v.(St).A != "A" || v.(St).B != 234 {
		t.Error("bad value")
	}
	v, err = parseJsonParam(strings.NewReader(`{"A":"111"}`), &st)
	if err != nil {
		t.Error(err)
	}
	if v.(*St).A != "111" || v.(*St).B != 123 {
		t.Error("bad value")
	}

	//httptest.NewServer().Close()
}

func TestNewFuncHandler(t *testing.T) {
	handler := NewFuncHandler(func(n int) int {
		return n * 2
	})
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte("7")))
	if err != nil {
		t.Error(err)
	}
	handler.ServeHTTP(resp, req)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	envelop := OcpAgentResponse{
		Data: 0,
	}
	err = json.Unmarshal(body, &envelop)
	if err != nil {
		t.Error(err)
	}
	if envelop.Data != 14 {
		t.Error("bad value")
	}
	fmt.Println(string(body))
}
