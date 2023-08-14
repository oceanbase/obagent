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

package process

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/lib/system"
	"github.com/oceanbase/obagent/tests/mock"
)

func TestProcessExists(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockProcess := mock.NewMockProcess(ctl)
	libProcess = mockProcess

	type args struct {
		name   string
		exists bool
	}
	type want struct {
		exists bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "oceanbase process exists",
			args: args{name: "oceanbase", exists: true},
			want: want{exists: true},
		},
		{
			name: "test process not exists",
			args: args{name: "test", exists: false},
			want: want{exists: false},
		},
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockProcess.EXPECT().ProcessExists(tt.args.name).Return(tt.args.exists, nil)

			exists, err := ProcessExists(context.Background(), CheckProcessExistsParam{Name: tt.args.name})
			So(err, ShouldBeNil)
			So(exists, ShouldEqual, tt.want.exists)
		})
	}
}

func TestGetProcessInfo(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockProcess := mock.NewMockProcess(ctl)
	libProcess = mockProcess

	Convey("get process info", t, func() {
		processName := "oceanbase"

		mockProcess.EXPECT().FindProcessInfoByName(processName).Return([]*system.ProcessInfo{
			{Pid: 1001, Name: processName},
			{Pid: 1002, Name: processName},
		}, nil)

		processInfoResult, err := GetProcessInfo(context.Background(), GetProcessInfoParam{Name: processName})
		So(err, ShouldBeNil)
		processInfos := processInfoResult.ProcessInfoList
		So(processInfos, ShouldHaveLength, 2)
		So(processInfos[0].Name, ShouldEqual, processName)
		So(processInfos[1].Name, ShouldEqual, processName)
	})
}

func TestStopProcess(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockProcess := mock.NewMockProcess(ctl)
	libProcess = mockProcess

	type args struct {
		name  string
		force bool
	}
	type want struct {
		force bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "no force stop",
			args: args{name: "oceanbase", force: false},
			want: want{force: false},
		},
		{
			name: "force stop",
			args: args{name: "oceanbase", force: true},
			want: want{force: true},
		},
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			param := StopProcessParam{
				Process: FindProcessParam{
					FindType: byName,
					Name:     tt.args.name,
				},
				Force: tt.args.force,
			}

			if tt.want.force {
				mockProcess.EXPECT().KillProcessByName(tt.args.name).Return(nil)
			} else {
				mockProcess.EXPECT().TerminateProcessByName(tt.args.name).Return(nil)
			}

			err := StopProcess(context.Background(), param)
			So(err, ShouldBeNil)
		})
	}
}
