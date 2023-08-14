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

package command

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	store := NewFileTaskStore(os.TempDir())
	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	token := ExecutionToken{id}
	execution := structuredExecution()
	err := store.Create(token, execution)
	if err != nil {
		t.Error("first create should success")
	}
	err = store.Create(token, execution)
	if err == nil {
		t.Error("second create should fail")
	}
}

func TestStoreLoadStructured(t *testing.T) {
	store := NewFileTaskStore(os.TempDir())
	token := ExecutionToken{"123"}
	execution := structuredExecution()
	err := store.Store(token, execution)
	if err != nil {
		t.Error("store failed", err)
		return
	}
	status, err := store.Load(token)
	if err != nil {
		t.Error("load failed", err)
	}
	fmt.Printf("%+v\n", status)
	if status.ResponseType != TypeStructured {
		t.Error("bad ResponseType")
	}
	if status.Param == nil || status.Annotation == nil || status.Finished == false || status.Ok == false || status.Result == nil {
		t.Error("bad status")
	}
	_ = store.Delete(token)
}

func structuredExecution() *Execution {
	input := &Input{
		annotation: map[string]interface{}{
			"command": "testCommand",
		},
		param: map[string]interface{}{
			"arg0": 123,
			"arg1": "xxx",
		},
	}
	output := &Output{
		result: map[string]interface{}{
			"a": 1,
			"b": "2",
			"c": []int{3, 3, 3},
		},
		err:      nil,
		finished: true,
		ok:       true,
	}
	ctx := &ExecutionContext{
		input:  input,
		output: output,
	}
	execution := &Execution{
		cmd: &Task{
			dataType: TypeStructured,
		},
		ctx: ctx,
	}
	return execution
}

func TestStoreLoadText(t *testing.T) {
	store := NewFileTaskStore(os.TempDir())
	token := ExecutionToken{"123"}
	fmt.Println(store.Path(token))
	input := &Input{
		annotation: map[string]interface{}{
			"command": "testCommand",
		},
		param: map[string]interface{}{
			"arg0": 123,
			"arg1": "xxx",
		},
	}
	output := &Output{
		result:   "result text",
		err:      nil,
		finished: true,
		ok:       true,
	}
	ctx := &ExecutionContext{
		input:  input,
		output: output,
	}
	execution := &Execution{
		cmd: &Task{
			dataType: TypeText,
		},
		ctx: ctx,
	}
	err := store.Store(token, execution)
	if err != nil {
		t.Error("store failed", err)
		return
	}
	status, err := store.Load(token)
	if err != nil {
		t.Error("load failed", err)
	}
	fmt.Printf("%+v\n", status)
	if status.ResponseType != TypeText {
		t.Error("bad ResponseType")
	}
	if status.Param == nil || status.Annotation == nil || status.Finished == false || status.Ok == false || status.Result == nil {
		t.Error("bad status")
	}
	_ = store.Delete(token)
}

func TestStoreLoadBinary(t *testing.T) {
	store := NewFileTaskStore(os.TempDir())
	token := ExecutionToken{"123"}
	fmt.Println(store.Path(token))
	input := &Input{
		annotation: map[string]interface{}{
			"command": "testCommand",
		},
		param: map[string]interface{}{
			"arg0": 123,
			"arg1": "xxx",
		},
	}
	output := &Output{
		result:   []byte("result binary"),
		err:      nil,
		finished: true,
		ok:       true,
	}
	ctx := &ExecutionContext{
		input:  input,
		output: output,
	}
	execution := &Execution{
		cmd: &Task{
			dataType: TypeBinary,
		},
		ctx: ctx,
	}
	err := store.Store(token, execution)
	if err != nil {
		t.Error("store failed", err)
		return
	}
	status, err := store.Load(token)
	if err != nil {
		t.Error("load failed", err)
	}
	fmt.Printf("%+v\n", status)
	if status.ResponseType != TypeBinary {
		t.Error("bad ResponseType")
	}
	if status.Param == nil || status.Annotation == nil || status.Finished == false || status.Ok == false || status.Result == nil {
		t.Error("bad status")
	}
	_ = store.Delete(token)
}

func TestCleanup(t *testing.T) {
	store := NewFileTaskStore(os.TempDir())
	store.Cleanup(time.Minute)
}
