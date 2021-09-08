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

package server

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTcpListener(t *testing.T) {
	listener1, err := NewTcpListener("127.0.0.1:9998")
	if err != nil {
		t.Error("run tcp listener failed", err)
	}
	defer listener1.Close()
	listener2, err := NewTcpListener("127.0.0.1:9998")
	if err != nil {
		t.Error("run tcp listener on same port failed", err)
	}
	defer listener2.Close()
}

func TestStartTcp(t *testing.T) {
	defer time.Sleep(time.Millisecond)

	listener1 := NewListener()
	err := listener1.StartTCP("127.0.0.1:9998")
	if err != nil {
		t.Error("run StartTCP failed", err)
	}
	defer listener1.Close()

	listener2 := NewListener()
	err = listener2.StartTCP("127.0.0.1:9998")
	if err != nil {
		t.Error("run StartTCP failed", err)
	}
	defer listener2.Close()

}

func TestStartErr(t *testing.T) {
	defer time.Sleep(time.Millisecond)

	listener1 := NewListener()
	err := listener1.StartTCP("127.0.0.11:9998")
	if err == nil {
		fmt.Println("run StartTCP on bad address should failed")
		//t.Error("run StartTCP on bad address should failed")
		defer listener1.Close()
	}
	fmt.Println(err)

	err = listener1.StartSocket("/not/exist_file")
	if err == nil {
		t.Error("run StartTCP on bad address should failed")
		defer listener1.Close()
	}
	fmt.Println(err)
}
