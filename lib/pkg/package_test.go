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

package pkg

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/lib/shell"
	"github.com/oceanbase/obagent/lib/system"
	"github.com/oceanbase/obagent/tests/mock"
)

func TestPackageInstalled(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockShellf := mock.NewMockShellShelf(ctl)
	mockCommand := mock.NewMockCommand(ctl)
	libShellf = mockShellf

	type args struct {
		packageCount int
		packageName  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "one package info",
			args: args{
				packageCount: 1,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
		},
		{
			name: "none package info",
			args: args{
				packageCount: 0,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
		},
		{
			name: "multiple package info",
			args: args{
				packageCount: 2,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
		},
	}

	libPackage := PackageImpl{}
	for _, tt := range tests {
		output := ""
		for i := 0; i < tt.args.packageCount; i++ {
			output += tt.args.packageName
			output += "\n"
		}
		Convey(tt.name, t, func() {
			mockCommand.EXPECT().WithUser(gomock.Any()).AnyTimes().Return(mockCommand)
			mockShellf.EXPECT().GetCommandForCurrentPlatform(gomock.Any(), gomock.Any()).Return(mockCommand, nil)
			mockCommand.EXPECT().ExecuteAllowFailure().Return(&shell.ExecuteResult{
				ExitCode: 0,
				Output:   output,
			}, nil)

			packageInstalled, err := libPackage.PackageInstalled("oceanbase")
			So(err, ShouldBeNil)
			So(packageInstalled, ShouldEqual, tt.args.packageCount > 0)
		})
	}
}

func TestGetPackageInfo(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockShellf := mock.NewMockShellShelf(ctl)
	mockCommand := mock.NewMockCommand(ctl)
	libShellf = mockShellf

	type args struct {
		packageCount int
		packageName  string
	}
	type want struct {
		name         string
		version      string
		buildNumber  string
		os           string
		architecture string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "one package info",
			args: args{
				packageCount: 1,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
			want: want{
				name:         "oceanbase",
				version:      "2.2.77",
				buildNumber:  "20210522122736",
				os:           "el7",
				architecture: "x86_64",
			},
		},
		{
			name: "none package info",
			args: args{
				packageCount: 0,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
			want: want{
				name:         "oceanbase",
				version:      "2.2.77",
				buildNumber:  "20210522122736",
				os:           "el7",
				architecture: "x86_64",
			},
		},
		{
			name: "multiple package info",
			args: args{
				packageCount: 2,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
			want: want{
				name:         "oceanbase",
				version:      "2.2.77",
				buildNumber:  "20210522122736",
				os:           "el7",
				architecture: "x86_64",
			},
		},
	}

	libPackage := PackageImpl{}
	for _, tt := range tests {
		output := ""
		for i := 0; i < tt.args.packageCount; i++ {
			output += tt.args.packageName
			output += "\n"
		}
		Convey(tt.name, t, func() {
			mockCommand.EXPECT().WithUser(gomock.Any()).AnyTimes().Return(mockCommand)
			mockShellf.EXPECT().GetCommandForCurrentPlatform(gomock.Any(), gomock.Any()).Return(mockCommand, nil)
			mockCommand.EXPECT().ExecuteAllowFailure().Return(&shell.ExecuteResult{
				ExitCode: 0,
				Output:   output,
			}, nil)

			packageInfo, err := libPackage.GetPackageInfo("oceanbase")
			if tt.args.packageCount == 1 {
				So(err, ShouldBeNil)
				So(packageInfo.Name, ShouldEqual, tt.want.name)
				So(packageInfo.Version, ShouldEqual, tt.want.version)
				So(packageInfo.BuildNumber, ShouldEqual, tt.want.buildNumber)
				So(packageInfo.Os, ShouldEqual, tt.want.os)
				So(packageInfo.Architecture, ShouldEqual, tt.want.architecture)
			} else {
				So(err, ShouldNotBeNil)
			}
		})
	}
}

func TestFindPackageInfo(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockShellf := mock.NewMockShellShelf(ctl)
	mockCommand := mock.NewMockCommand(ctl)
	libShellf = mockShellf

	type args struct {
		packageCount int
		packageName  string
	}
	type want struct {
		name         string
		version      string
		buildNumber  string
		os           string
		architecture string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "one package info",
			args: args{
				packageCount: 1,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
			want: want{
				name:         "oceanbase",
				version:      "2.2.77",
				buildNumber:  "20210522122736",
				os:           "el7",
				architecture: "x86_64",
			},
		},
		{
			name: "none package info",
			args: args{
				packageCount: 0,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
			want: want{
				name:         "oceanbase",
				version:      "2.2.77",
				buildNumber:  "20210522122736",
				os:           "el7",
				architecture: "x86_64",
			},
		},
		{
			name: "multiple package info",
			args: args{
				packageCount: 2,
				packageName:  "oceanbase-2.2.77-20210522122736.el7.x86_64",
			},
			want: want{
				name:         "oceanbase",
				version:      "2.2.77",
				buildNumber:  "20210522122736",
				os:           "el7",
				architecture: "x86_64",
			},
		},
	}

	libPackage := PackageImpl{}
	for _, tt := range tests {
		output := ""
		for i := 0; i < tt.args.packageCount; i++ {
			output += tt.args.packageName
			output += "\n"
		}
		Convey(tt.name, t, func() {
			mockCommand.EXPECT().WithUser(gomock.Any()).AnyTimes().Return(mockCommand)
			mockShellf.EXPECT().GetCommandForCurrentPlatform(gomock.Any(), gomock.Any()).Return(mockCommand, nil)
			mockCommand.EXPECT().ExecuteAllowFailure().Return(&shell.ExecuteResult{
				ExitCode: 0,
				Output:   output,
			}, nil)

			packageInfoList, err := libPackage.FindPackageInfo("oceanbase")
			So(err, ShouldBeNil)
			So(len(packageInfoList), ShouldEqual, tt.args.packageCount)
			for _, packageInfo := range packageInfoList {
				So(packageInfo.FullPackageName, ShouldEqual, tt.args.packageName)
				So(packageInfo.Name, ShouldEqual, tt.want.name)
				So(packageInfo.Version, ShouldEqual, tt.want.version)
				So(packageInfo.BuildNumber, ShouldEqual, tt.want.buildNumber)
				So(packageInfo.Os, ShouldEqual, tt.want.os)
				So(packageInfo.Architecture, ShouldEqual, tt.want.architecture)
			}
		})
	}
}

func TestInstallPackage(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockShellf := mock.NewMockShellShelf(ctl)
	mockCommand := mock.NewMockCommand(ctl)
	libShellf = mockShellf

	mockSystem := mock.NewMockSystem(ctl)
	libSystem = mockSystem

	libPackage := PackageImpl{}
	Convey("install package", t, func() {
		mockSystem.EXPECT().GetHostInfo().Return(&system.HostInfo{OsPlatformFamily: ""}, nil)
		mockShellf.EXPECT().GetCommandForCurrentPlatform(gomock.Any(), gomock.Any()).Return(mockCommand, nil)
		mockCommand.EXPECT().WithUser(gomock.Any()).AnyTimes().Return(mockCommand)
		mockCommand.EXPECT().WithTimeout(gomock.Any()).AnyTimes().Return(mockCommand)
		mockCommand.EXPECT().Execute().Return(mockExecuteResult(true), nil)

		err := libPackage.InstallPackage("/tmp/rpms/oceanbase-2.2.77-20210522122736.el7.x86_64.rpm")
		So(err, ShouldBeNil)
	})
}

func TestUninstallPackage(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockShellf := mock.NewMockShellShelf(ctl)
	mockCommand := mock.NewMockCommand(ctl)
	libShellf = mockShellf

	libPackage := PackageImpl{}
	Convey("uninstall package", t, func() {
		mockShellf.EXPECT().GetCommandForCurrentPlatform(gomock.Any(), gomock.Any()).Return(mockCommand, nil)
		mockCommand.EXPECT().WithUser(gomock.Any()).AnyTimes().Return(mockCommand)
		mockCommand.EXPECT().Execute().Return(mockExecuteResult(true), nil)

		err := libPackage.UninstallPackage("oceanbase")
		So(err, ShouldBeNil)
	})
}

func mockExecuteResult(successful bool) *shell.ExecuteResult {
	if successful {
		return successfulExecuteResult()
	} else {
		return failedExecuteResult()
	}
}

func successfulExecuteResult() *shell.ExecuteResult {
	return &shell.ExecuteResult{ExitCode: 0, Output: "mock data\n"}
}

func failedExecuteResult() *shell.ExecuteResult {
	return &shell.ExecuteResult{ExitCode: 1}
}

func Test_parsePackageInfo(t *testing.T) {
	_, err := parsePackageInfo("dpkg: warning: failed to open configuration file '/root/.dpkg.cfg' for reading: Permission denied")
	if err == nil {
		t.Errorf("invalid package should fail")
	}
	info, err := parsePackageInfo("obproxy-1.8.6-20210210153306.el7.x86_64")
	if err != nil {
		t.Errorf("should success")
		return
	}
	if info.Name != "obproxy" || info.Version != "1.8.6" || info.BuildNumber != "20210210153306" || info.Os != "el7" || info.Architecture != "x86_64" {
		t.Errorf("package info wrong %v", info)
	}
}
