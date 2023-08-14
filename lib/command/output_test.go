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
	"errors"
	"testing"
	"time"
)

func TestOutput_Init(t *testing.T) {
	out := newOutput()
	if out.Finished() == true {
		t.Error("initial output MUST NOT be finished")
	}
}

func TestOutput_FinishOk(t *testing.T) {
	out := newOutput()

	err := out.FinishOk("result")
	if err != nil {
		t.Error("finish failed")
	}
	if !out.Finished() {
		t.Error("should be finished")
	}
	if !out.Ok() {
		t.Error("should be ok")
	}

	result, err := out.Get()
	if err != nil {
		t.Error("get result should success")
	}
	if result != "result" {
		t.Error("result wrong")
	}
}

func TestOutput_FinishErr(t *testing.T) {
	out := newOutput()

	err := out.FinishErr(errors.New("errors"))
	if err != nil {
		t.Error("finish failed")
	}
	if !out.Finished() {
		t.Error("should be finished")
	}
	if out.Ok() {
		t.Error("should not be ok")
	}

	_, err = out.Get()
	if err == nil {
		t.Error("get should return errors")
		return
	}
	if err.Error() != "errors" {
		t.Error("errors wrong")
	}
}

func TestOutput_Wait(t *testing.T) {
	out := newOutput()
	go func() {
		time.Sleep(time.Millisecond * 200)
		out.FinishOk("ok")
	}()
	result, err := out.Get()
	if err != nil || result != "ok" {
		t.Error("bad result")
	}
}

func TestOutput_WaitTimeout(t *testing.T) {
	out := newOutput()
	go func() {
		time.Sleep(time.Millisecond * 100)
		out.FinishOk("ok")
	}()
	result, err := out.GetWithTimeout(time.Millisecond * 200)
	if err != nil || result != "ok" {
		t.Error("bad result")
	}
}

func TestOutput_WaitTimeout2(t *testing.T) {
	out := newOutput()
	go func() {
		time.Sleep(time.Second)
		out.FinishOk("ok")
	}()
	_, err := out.GetWithTimeout(time.Millisecond * 100)
	if err == nil || err != TimeoutErr {
		t.Error("should be timeout")
	}
}

func TestOutput_Chan(t *testing.T) {
	out := newOutput()
	go func() {
		time.Sleep(time.Millisecond * 200)
		out.FinishOk("ok")
	}()
	select {
	case <-out.Done():
	case <-time.After(time.Second):
		t.Error("select wait chan timeout")
		return
	}
	if !out.Finished() {
		t.Error("should be finished")
	}
}

func TestOutput_Progress(t *testing.T) {
	out := newOutput()
	out.UpdateProgress(1)
	if out.Progress() != 1 {
		t.Error("progress should be 1")
	}
	go func() {
		out.UpdateProgress(2)
		time.Sleep(time.Millisecond * 200)
		out.FinishOk("ok")
	}()
	status := out.Status()
	if status.Finished == true || status.Progress == nil {
		t.Error("bad status in progress")
	}
	_, _ = out.Get()
	if out.Progress() != 2 {
		t.Error("progress should be 2")
	}
	status = out.Status()
	if status.Finished == false || status.Progress != 2 || !status.Ok || status.Result != "ok" {
		t.Error("bad status after done")
	}
}

func TestOutput_DoubleFinish(t *testing.T) {
	out := newOutput()
	out.FinishOk("ok")
	err := out.FinishErr(errors.New("errors"))
	if err == nil {
		t.Error("double finish should failed")
	}
	if !out.Ok() {
		t.Error("double finish should not change status")
	}

	out2 := newOutput()
	out2.FinishErr(errors.New("errors"))
	err = out2.FinishOk("ok")
	if err == nil {
		t.Error("double finish should failed")
	}
	if out2.Ok() {
		t.Error("double finish should not change status")
	}
}
