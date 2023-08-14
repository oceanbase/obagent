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

package cleaner

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func prepareTestDirTree(tree string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v\n", err)
	}

	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		return "", fmt.Errorf("error evaling temp directory: %v\n", err)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, tree), 0755)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return filepath.Join(tmpDir, tree), nil
}

func TestGetRealPath(t *testing.T) {
	tmpDirA, err := prepareTestDirTree("a")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDirA)

	tmpDirB, err := prepareTestDirTree("b")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDirB)
	tmpDirC := filepath.Join(tmpDirB, "c")

	err = os.Symlink(tmpDirA, tmpDirC)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDirC)

	tests := []struct {
		description  string
		symbolicLink string
		realPath     string
		wantErr      bool
	}{
		{
			description:  "普通路径，没有软链的情况",
			realPath:     tmpDirA,
			symbolicLink: tmpDirA,
		},
		{
			description:  "路径为软链的情况",
			symbolicLink: tmpDirC,
			realPath:     tmpDirA,
		},
		{
			description:  "路径不存在",
			symbolicLink: "empty",
			realPath:     "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		Convey(tt.description, t, func() {
			symlinkRealPath, err := GetRealPath(context.Background(), tt.symbolicLink)
			if !tt.wantErr {
				So(err, ShouldBeNil)
				So(symlinkRealPath, ShouldEqual, tt.realPath)
			} else {
				So(err, ShouldNotBeNil)
			}
		})
	}
}

func TestGetDiskUsage(t *testing.T) {
	tmpDir, err := prepareTestDirTree("tmp1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	tmpFile, err := os.OpenFile("test.txt", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tmpFile.WriteString("Test GetDiskUsage ~~~")
	if err != nil {
		t.Fatal(err)
	}

	Convey("查询磁盘使用率", t, func() {
		percentage, err := GetDiskUsage(context.Background(), tmpDir)
		So(err, ShouldBeNil)
		So(percentage, ShouldNotBeZeroValue)
	})

	Convey("查询不存在的路径", t, func() {
		percentage, err := GetDiskUsage(context.Background(), "empty")
		So(err, ShouldNotBeNil)
		So(percentage, ShouldBeZeroValue)
	})
}

func TestFindFilesByRegexAndMTime(t *testing.T) {
	tmpDir, err := prepareTestDirTree("tmp1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	filesToBeCreated := []string{"a.b.log", "a.b.log.1", "b.log.2", "c.d.log.wf.1", "log.wf.1", "f.log.wf.2"}

	for _, filesToBeCreated := range filesToBeCreated {
		_, err := os.OpenFile(filesToBeCreated, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	now := time.Now()
	err = os.Chtimes("a.b.log.1", now, now.AddDate(0, 0, -2))
	if err != nil {
		t.Fatal()
	}

	type args struct {
		dir     string
		fileReg string
		mTime   int64
	}
	tests := []struct {
		description  string
		args         args
		matchedFiles []string
	}{
		{
			description: "查询 1 天内满足给定 regex 的文件",
			args: args{
				dir:     tmpDir,
				fileReg: "([a-z]+.)?[a-z]+.log.[0-9]+",
				mTime:   -1,
			},
			matchedFiles: []string{"b.log.2"},
		},
		{
			description: "查询 1 天前满足给定 regex 的文件",
			args: args{
				dir:     tmpDir,
				fileReg: "([a-z]+.)?[a-z]+.log.[0-9]+",
				mTime:   1,
			},
			matchedFiles: []string{"a.b.log.1"},
		},
		{
			description: "查询 1 天内满足给定 regex 的文件",
			args: args{
				dir:     tmpDir,
				fileReg: "([a-z]+.)?[a-z]+.log.wf.[0-9]+",
				mTime:   -1,
			},
			matchedFiles: []string{"c.d.log.wf.1", "f.log.wf.2"},
		},
	}
	for _, tt := range tests {
		Convey(tt.description, t, func() {
			mTimeDuration := time.Duration(tt.args.mTime) * time.Hour * 24
			matchedFiles, err := FindFilesByRegexAndMTime(context.Background(), tt.args.dir, tt.args.fileReg, mTimeDuration)
			So(err, ShouldBeNil)
			for i := range matchedFiles {
				fmt.Println(matchedFiles[i].Path)
				So(matchedFiles[i].Path, ShouldEqual, filepath.Join(tmpDir, tt.matchedFiles[i]))
			}
		})
	}
}

func TestFindFilesAndSortByMTime(t *testing.T) {
	tmpDir, err := prepareTestDirTree("tmp1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	filesToBeCreated := []string{"a.b.log", "a.b.log.1", "b.log.2", "c.d.log.wf.1", "log.wf.1", "f.log.wf.2"}

	for _, fileToBeCreated := range filesToBeCreated {
		_, err := os.OpenFile(fileToBeCreated, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	now := time.Now()
	err = os.Chtimes("f.log.wf.2", now, now.AddDate(0, 0, -2))
	if err != nil {
		t.Fatal()
	}
	err = os.Chtimes("a.b.log.1", now, now.AddDate(0, 0, -2))
	if err != nil {
		t.Fatal()
	}
	type args struct {
		dir     string
		fileReg string
	}
	tests := []struct {
		description  string
		args         args
		matchedFiles []string
	}{
		{
			description: "查询所有满足给定 regex 的文件，且返回的文件按照 MTime 排序",
			args: args{
				dir:     tmpDir,
				fileReg: "([a-z]+.)?[a-z]+.log.[0-9]+",
			},
			matchedFiles: []string{"a.b.log.1", "b.log.2"},
		}, {
			description: "查询所有满足给定 regex 的文件，且返回的文件按照 MTime 排序",
			args: args{
				dir:     tmpDir,
				fileReg: "([a-z]+.)?[a-z]+.log.wf.[0-9]+",
			},
			matchedFiles: []string{"f.log.wf.2", "c.d.log.wf.1"},
		}, {
			description: "fileRegex 参数为空字符串，默认匹配所有",
			args: args{
				dir:     tmpDir,
				fileReg: "",
			},
			matchedFiles: []string{"a.b.log.1", "f.log.wf.2", "a.b.log", "b.log.2", "c.d.log.wf.1", "log.wf.1"},
		},
	}
	for _, tt := range tests {
		Convey(tt.description, t, func() {
			matchedFiles, err := FindFilesAndSortByMTime(context.Background(), tt.args.dir, tt.args.fileReg)
			So(err, ShouldBeNil)
			for i := range matchedFiles {
				So(matchedFiles[i].Path, ShouldEqual, filepath.Join(tmpDir, tt.matchedFiles[i]))
			}
		})
	}
}

func Test_matchRegex(t *testing.T) {
	type args struct {
		reg     string
		content string
	}
	tests := []struct {
		name        string
		args        args
		wantMatched bool
		wantErr     bool
	}{
		{
			name: "normal observer.log",
			args: args{
				reg:     "([a-z]+.)?[a-z]+.log.[0-9]+",
				content: "observer.log.20210824131916",
			},
			wantMatched: true,
			wantErr:     false,
		},
		{
			name: "abnormal observer.log",
			args: args{
				reg:     "([a-z]+.)?[a-z]+.log.[0-9]+",
				content: "observer.log",
			},
			wantMatched: false,
			wantErr:     false,
		},
		{
			name: "normal observer.log.wf",
			args: args{
				reg:     "([a-z]+.)?[a-z]+.log.wf.[0-9]+",
				content: "observer.log.wf.20210823083142",
			},
			wantMatched: true,
			wantErr:     false,
		},
		{
			name: "abnormal observer.log.wf",
			args: args{
				reg:     "([a-z]+.)?[a-z]+.log.wf.[0-9]+",
				content: "observer.log.wf",
			},
			wantMatched: false,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatched, err := matchRegex(tt.args.reg, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("matchRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatched != tt.wantMatched {
				t.Errorf("matchRegex() = %v, want %v", gotMatched, tt.wantMatched)
			}
		})
	}
}
