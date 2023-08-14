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

package file

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/shell"
	"github.com/oceanbase/obagent/tests/mock"
)

func TestDownloadFileFromUrl(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockHttp := mock.NewMockHttp(ctl)
	mockFile := mock.NewMockFile(ctl)
	libHttp = mockHttp
	libFile = mockFile

	type args struct {
		url              string
		target           string
		expectedChecksum string
		actualChecksum   string
	}
	type want struct {
		success bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "same checksum",
			args: args{
				url:              "http://127.0.0.1:8080/a.rpm",
				target:           "rpms/a.rpm",
				expectedChecksum: "2439573625385400f2a669657a7db6ae7515d371",
				actualChecksum:   "2439573625385400f2a669657a7db6ae7515d371",
			},
			want: want{
				success: true,
			},
		},
		{
			name: "different checksum",
			args: args{
				url:              "http://127.0.0.1:8080/b.rpm",
				target:           "rpms/b.rpm",
				expectedChecksum: "2439573625385400f2a669657a7db6ae7515d371",
				actualChecksum:   "b85e079d153ccb06cc01db22b074dba0e0ec0e26",
			},
			want: want{
				success: false,
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			fullPath := NewPathFromRelPath(tt.args.target).FullPath()
			mockFile.EXPECT().FileExists(gomock.Any()).Return(false, nil)
			mockFile.EXPECT().CreateDirectoryForUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockHttp.EXPECT().DownloadFile(fullPath, tt.args.url).Return(nil)
			mockFile.EXPECT().Sha1Checksum(fullPath).Return(tt.args.actualChecksum, nil)

			_, err := DownloadFile(context.Background(), DownloadFileParam{
				Source: DownloadFileSource{
					Url: tt.args.url,
				},
				Target:           tt.args.target,
				ValidateChecksum: true,
				Checksum:         tt.args.expectedChecksum,
			})
			if tt.want.success {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldNotBeNil)
			}
		})
	}
}

func TestDownloadFileFromLocalFile(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockHttp := mock.NewMockHttp(ctl)
	mockFile := mock.NewMockFile(ctl)
	libHttp = mockHttp
	libFile = mockFile

	type args struct {
		path   string
		target string
	}
	type want struct {
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "from local file",
			args: args{
				path:   "/home/admin/ocp_agent/lib",
				target: "rpms/a.rpm",
			},
			want: want{},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockFile.EXPECT().FileExists(gomock.Any()).Return(false, nil)
			mockFile.EXPECT().CreateDirectoryForUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			mockFile.EXPECT().CopyFile(tt.args.path, gomock.Any(), gomock.Any())

			_, err := DownloadFile(context.Background(), DownloadFileParam{
				Source: DownloadFileSource{
					Type: DownloadFileFromLocalFile,
					Path: tt.args.path,
				},
				Target:           tt.args.target,
				ValidateChecksum: false,
			})
			So(err, ShouldBeNil)
		})
	}
}

func TestIsFileExists(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockFile := mock.NewMockFile(ctl)
	libFile = mockFile

	type args struct {
		path   string
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
			name: "file exists",
			args: args{
				path:   "rpms/a.rpm",
				exists: true,
			},
			want: want{
				exists: true,
			},
		},
		{
			name: "file not exists",
			args: args{
				path:   "rpms/a.rpm",
				exists: true,
			},
			want: want{
				exists: true,
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockFile.EXPECT().FileExists(gomock.Any()).Return(tt.args.exists, nil)
			result, err := IsFileExists(context.Background(), GetFileExistsParam{
				FilePath: tt.args.path,
			})
			So(err, ShouldBeNil)
			So(result.Exists, ShouldEqual, tt.want.exists)
		})
	}
}

func TestRemoveFiles(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockFile := mock.NewMockFile(ctl)
	libFile = mockFile

	type args struct {
		path  string
		isDir bool
	}
	type want struct {
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "remove directory",
			args: args{
				path:  "rpms/a.rpm",
				isDir: true,
			},
			want: want{},
		},
		{
			name: "remove file",
			args: args{
				path:  "rpms",
				isDir: true,
			},
			want: want{},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockFile.EXPECT().IsDir(gomock.Any()).Return(tt.args.isDir)
			if tt.args.isDir {
				mockFile.EXPECT().RemoveDirectory(gomock.Any()).Return(nil)
			} else {
				mockFile.EXPECT().RemoveFileIfExists(gomock.Any()).Return(nil)
			}
			err := RemoveFiles(context.Background(), RemoveFilesParam{
				PathList: []string{tt.args.path},
			})
			So(err, ShouldBeNil)
		})
	}
}

func TestGetRealStaticPath(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockFile := mock.NewMockFile(ctl)
	libFile = mockFile

	type args struct {
		symLinkPath    string
		realStaticPath string
	}
	type want struct {
		realStaticPath string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "get real static path",
			args: args{
				symLinkPath:    "/home/admin/oceanbase/store",
				realStaticPath: "/data/1/cluster1",
			},
			want: want{
				realStaticPath: "/data/1/cluster1",
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockFile.EXPECT().GetRealPathBySymbolicLink(tt.args.symLinkPath).Return(tt.args.realStaticPath, nil)
			result, err := GetRealStaticPath(context.Background(), GetRealStaticPathParam{
				SymbolicLink: tt.args.symLinkPath,
			})
			So(err, ShouldBeNil)
			So(result.Path, ShouldEqual, tt.want.realStaticPath)
		})
	}
}

func TestCheckDirectoryPermission(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockFile := mock.NewMockFile(ctl)
	mockShell := mock.NewMockShell(ctl)
	mockCommand := mock.NewMockCommand(ctl)
	libFile = mockFile
	libShell = mockShell

	const (
		checkDirectoryExistsError = iota
		directoryNotExist
		pathIsNotDirectory
		checkDirectoryPermissionError
		directoryHasPermission
		directoryHasNoPermission
	)

	type args struct {
		scenario int
	}
	type want struct {
		checkResult CheckDirectoryPermissionResult
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "cannot check directory exists",
			args: args{
				scenario: checkDirectoryExistsError,
			},
			want: want{
				checkResult: checkFailed,
			},
		},
		{
			name: "directory not exists",
			args: args{
				scenario: directoryNotExist,
			},
			want: want{
				checkResult: directoryNotExists,
			},
		},
		{
			name: "path is not directory",
			args: args{
				scenario: pathIsNotDirectory,
			},
			want: want{
				checkResult: directoryNotExists,
			},
		},
		{
			name: "cannot check directory permission",
			args: args{
				scenario: checkDirectoryPermissionError,
			},
			want: want{
				checkResult: checkFailed,
			},
		},
		{
			name: "directory has permission",
			args: args{
				scenario: directoryHasPermission,
			},
			want: want{
				checkResult: hasPermission,
			},
		},
		{
			name: "directory has no permission",
			args: args{
				scenario: directoryHasNoPermission,
			},
			want: want{
				checkResult: noPermission,
			},
		},
	}

	ctx := context.Background()
	param := CheckDirectoryPermissionParm{
		Directory:  "/data/1",
		User:       "admin",
		Permission: accessWrite,
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			switch tt.args.scenario {
			case checkDirectoryExistsError:
				mockFile.EXPECT().FileExists(gomock.Any()).Return(false, errors.New("some error"))
			case directoryNotExist:
				mockFile.EXPECT().FileExists(gomock.Any()).Return(false, nil)
			case pathIsNotDirectory:
				mockFile.EXPECT().FileExists(gomock.Any()).Return(true, nil)
				mockFile.EXPECT().IsDir(gomock.Any()).Return(false)
			case checkDirectoryPermissionError:
				mockFile.EXPECT().FileExists(gomock.Any()).Return(true, nil)
				mockFile.EXPECT().IsDir(gomock.Any()).Return(true)
				mockShell.EXPECT().NewCommand(gomock.Any()).Return(mockCommand)
				mockCommand.EXPECT().WithContext(gomock.Any()).AnyTimes().Return(mockCommand)
				mockCommand.EXPECT().WithUser(gomock.Any()).AnyTimes().Return(mockCommand)
				mockCommand.EXPECT().WithOutputType(gomock.Any()).AnyTimes().Return(mockCommand)
				mockCommand.EXPECT().Execute().Return(failedExecuteResult(), failedExecuteResult().AsError())
			case directoryHasPermission:
				mockFile.EXPECT().FileExists(gomock.Any()).Return(true, nil)
				mockFile.EXPECT().IsDir(gomock.Any()).Return(true)
				mockShell.EXPECT().NewCommand(gomock.Any()).Return(mockCommand)
				mockCommand.EXPECT().WithContext(gomock.Any()).AnyTimes().Return(mockCommand)
				mockCommand.EXPECT().WithUser(gomock.Any()).AnyTimes().Return(mockCommand)
				mockCommand.EXPECT().Execute().Return(successfulExecuteResult("0\n"), nil)
			case directoryHasNoPermission:
				mockFile.EXPECT().FileExists(gomock.Any()).Return(true, nil)
				mockFile.EXPECT().IsDir(gomock.Any()).Return(true)
				mockShell.EXPECT().NewCommand(gomock.Any()).Return(mockCommand)
				mockCommand.EXPECT().WithContext(gomock.Any()).AnyTimes().Return(mockCommand)
				mockCommand.EXPECT().WithUser(gomock.Any()).AnyTimes().Return(mockCommand)
				mockCommand.EXPECT().Execute().Return(successfulExecuteResult("1\n"), nil)
			}

			result, err := CheckDirectoryPermission(ctx, param)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, tt.want.checkResult)
		})
	}
}

func TestGetDirectoryUsed(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockFile := mock.NewMockFile(ctl)
	libFile = mockFile

	type args struct {
		path      string
		usedBytes int64
	}
	type want struct {
		usedBytes int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "get directory used bytes",
			args: args{
				path:      "/data/1",
				usedBytes: 584930031,
			},
			want: want{
				usedBytes: 584930031,
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			mockFile.EXPECT().GetDirectoryUsedBytes(gomock.Any(), gomock.Any()).Return(tt.args.usedBytes, nil)
			result, err := GetDirectoryUsed(context.Background(), GetDirectoryUsedParam{
				Path: tt.args.path,
			})
			So(err, ShouldBeNil)
			So(result.DirectoryUsedBytes, ShouldEqual, tt.want.usedBytes)
		})
	}
}

func successfulExecuteResult(output string) *shell.ExecuteResult {
	return &shell.ExecuteResult{ExitCode: 0, Output: output}
}

func failedExecuteResult() *shell.ExecuteResult {
	return &shell.ExecuteResult{ExitCode: 1}
}
