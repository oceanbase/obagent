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

package pkg

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/lib/pkg"
	"github.com/oceanbase/obagent/lib/shell"
	"github.com/oceanbase/obagent/tests/mock"
	"github.com/oceanbase/obagent/tests/mock2"
)

func TestGetPackageInfo(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockPackage := mock2.NewMockPackage(ctl)
	libPackage = mockPackage

	type args struct {
		name         string
		version      string
		buildNumber  string
		os           string
		architecture string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "get oceanbase package info",
			args: args{
				name:         "oceanbase",
				version:      "2.2.77",
				buildNumber:  "20210522122736",
				os:           "el7",
				architecture: "x86_64",
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockPackage.EXPECT().FindPackageInfo(gomock.Any()).Return([]*pkg.PackageInfo{{
				Name:         tt.args.name,
				Version:      tt.args.version,
				BuildNumber:  tt.args.buildNumber,
				Os:           tt.args.os,
				Architecture: tt.args.architecture,
			}}, nil)
			packageInfos, err := GetPackageInfo(context.Background(), GetPackageInfoParam{Name: tt.args.name})
			So(err, ShouldBeNil)
			So(packageInfos.PackageInfoList, ShouldNotBeEmpty)
			packageInfo := packageInfos.PackageInfoList[0]
			So(packageInfo.Name, ShouldEqual, tt.args.name)
			So(packageInfo.Version, ShouldEqual, tt.args.version)
			So(packageInfo.BuildNumber, ShouldEqual, tt.args.buildNumber)
			So(packageInfo.Os, ShouldEqual, tt.args.os)
			So(packageInfo.Architecture, ShouldEqual, tt.args.architecture)
		})
	}
}

func TestInstallPackage(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockPackage := mock2.NewMockPackage(ctl)
	libPackage = mockPackage

	type args struct {
		name          string
		alreadyExists bool
		customPath    bool
	}
	type want struct {
		changed bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "install package skipped",
			args: args{
				name:          "oceanbase",
				alreadyExists: true,
			},
			want: want{changed: false},
		},
		{
			name: "install package done",
			args: args{name: "test", alreadyExists: false},
			want: want{changed: true},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockPackage.EXPECT().PackageInstalled(gomock.Any()).Return(tt.args.alreadyExists, nil)
			if !tt.args.alreadyExists {
				mockPackage.EXPECT().InstallPackage(gomock.Any()).Return(nil)
			}
			mockPackage.EXPECT().GetPackageInfo(gomock.Any()).Return(&pkg.PackageInfo{}, nil)

			installPackageResult, err := InstallPackage(context.Background(), InstallPackageParam{Name: tt.args.name})
			So(err, ShouldBeNil)
			So(installPackageResult.Changed, ShouldEqual, tt.want.changed)
		})
	}
}

func TestUninstallPackage(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockPackage := mock2.NewMockPackage(ctl)
	libPackage = mockPackage

	type args struct {
		name   string
		exists bool
	}
	type want struct {
		changed bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "uninstall package skipped",
			args: args{name: "oceanbase", exists: false},
			want: want{changed: false},
		},
		{
			name: "uninstall package done",
			args: args{name: "test", exists: true},
			want: want{changed: true},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockPackage.EXPECT().PackageInstalled(gomock.Any()).Return(tt.args.exists, nil)
			if tt.args.exists {
				mockPackage.EXPECT().UninstallPackage(gomock.Any()).Return(nil)
			}

			uninstallPackageResult, err := UninstallPackage(context.Background(), UninstallPackageParam{Name: tt.args.name})
			So(err, ShouldBeNil)
			So(uninstallPackageResult.Changed, ShouldEqual, tt.want.changed)
		})
	}
}

func TestExtractPackage(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockFile := mock.NewMockFile(ctl)
	libFile = mockFile
	mockPackage := mock2.NewMockPackage(ctl)
	libPackage = mockPackage

	type args struct {
		extractAll bool
		rpmFile    string
		targetPath string
		fileInRpm  string
	}
	type want struct {
		changed bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "extract all files",
			args: args{
				extractAll: true,
				rpmFile:    "rpms/oceanbase-2.2.77-20210522122736.el7.x86_64.rpm",
				targetPath: "rpms/extract",
			},
			want: want{changed: false},
		},
		{
			name: "extract single file",
			args: args{
				extractAll: false,
				rpmFile:    "rpms/oceanbase-2.2.77-20210522122736.el7.x86_64.rpm",
				targetPath: "rpms/extract",
				fileInRpm:  "/home/admin/oceanbase/etc/upgrade_pre.py",
			},
			want: want{changed: true},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockFile.EXPECT().RemoveDirectory(gomock.Any()).Return(nil)
			mockFile.EXPECT().CreateDirectoryForUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			if tt.args.extractAll {
				mockPackage.EXPECT().ExtractPackageAllFiles(gomock.Any(), gomock.Any()).Return(nil)
			} else {
				mockPackage.EXPECT().ExtractPackageSingleFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			}

			var param ExtractPackageParam
			if tt.args.extractAll {
				param = ExtractPackageParam{
					ExtractAll:  true,
					PackageFile: tt.args.rpmFile,
					TargetPath:  tt.args.targetPath,
				}
			} else {
				param = ExtractPackageParam{
					ExtractAll:    false,
					PackageFile:   tt.args.rpmFile,
					TargetPath:    tt.args.targetPath,
					FileInPackage: tt.args.fileInRpm,
				}
			}
			extractPackageResult, err := ExtractPackage(context.Background(), param)
			So(err, ShouldBeNil)
			fmt.Printf("%#v\n", extractPackageResult)
		})
	}
}

func successfulExecuteResult() *shell.ExecuteResult {
	return &shell.ExecuteResult{ExitCode: 0}
}
