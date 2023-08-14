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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type Struct struct {
	Value string
}

func TestClient(t *testing.T) {
	listener := NewListener()
	listener.AddHandler("/hello", NewFuncHandler(func(s string) Struct {
		return Struct{Value: "hello " + s}
	}))
	listener.AddHandler("/error", NewFuncHandler(func() error {
		return errors.New("error")
	}))
	defer listener.Close()
	socket := filepath.Join(os.TempDir(), fmt.Sprintf("test_%d.sock", time.Now().UnixNano()))
	err := listener.StartSocket(socket)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(socket)

	if !CanConnect("unix", socket, time.Second) {
		t.Error("CanConnect should return true")
	}

	client := NewSocketApiClient(socket, time.Second)
	ret := &Struct{}
	err = client.Call("hello", "world", ret)
	if err != nil {
		t.Error(err)
	}
	if ret.Value != "hello world" {
		t.Error("parse result wrong")
	}

	err = client.Call("/not_exists", nil, nil)
	if err == nil {
		t.Error("should error on not_exists api")
	}
	fmt.Println(err)

	err = client.Call("/error", nil, nil)
	if err == nil {
		t.Error("should error on not_exists api")
	}
	fmt.Println(err)
}
