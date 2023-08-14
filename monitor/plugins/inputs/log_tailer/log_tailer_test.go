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

package log_tailer

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/message"
)

const logAtLayout = "2006-01-02 15:04:05.000000"
const logTimeInFileNameLayout = "20060102150405"

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

func TestLogTailer_checkAndOpenFile(t *testing.T) {
	tmpDir, err := prepareTestDirTree("tmp")
	if err != nil {
		t.Fatal(err)
	}

	tmpFile, err := os.Create(fmt.Sprintf("%s/%s", tmpDir, "ob.log.wf"))
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		logFileRealPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *os.File
		wantErr bool
	}{
		{
			name: "文件存在，返回一个 fd",
			args: args{
				logFileRealPath: tmpFile.Name(),
			},
			want:    tmpFile,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			fd, err := checkAndOpenFile(context.Background(), tt.args.logFileRealPath)
			if !tt.wantErr {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldNotBeNil)
				So(fd.Name(), ShouldEqual, tt.want.Name())
			}
		})
	}
}

func writeFile(tmpFile *os.File, start, end int) (*os.File, error) {
	for i := start; i < end; i++ {
		_, err := tmpFile.WriteString(fmt.Sprintf("[%s] ERROR  [RS] ob_server_table_operator.cpp:376 test %d ret=-%d\n",
			time.Now().Add(-2*time.Minute).Format(logAtLayout), i, i))
		if err != nil {
			return nil, err
		}
	}
	_, err := tmpFile.WriteString(fmt.Sprintf("[%s] ERROR  [RS] ob_server_table_operator.cpp:376 test matchFilter ret=-999\n",
		time.Now().Add(-2*time.Minute).Format(logAtLayout)))
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}

func createAndWriteFile(tmpDir string, fileName string, start, end int) (tmpFile *os.File, err error) {
	tmpFile, err = os.Create(fmt.Sprintf("%s/%s", tmpDir, fileName))
	if err != nil {
		return
	}
	return writeFile(tmpFile, start, end)
}

func TestLogTailer_handleLogFiles(t *testing.T) {
	invocationCnt := int64(0)
	totalCnt := 10

	tmpDir, err := prepareTestDirTree("tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile, err := createAndWriteFile(tmpDir, "ob.log.wf", 0, totalCnt)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	tmpReadFD, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer tmpReadFD.Close()

	type args struct {
		initLogQueue []*logFileInfo
		tailConf     monagent.TailConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "处理文件日志",
			args: args{
				initLogQueue: []*logFileInfo{
					{
						fileDesc:        tmpReadFD,
						isRenamed:       false,
						logAnalyzerType: "ob_light",
					},
				},
				tailConf: monagent.TailConfig{
					ProcessLogInterval: time.Second,
					LogSourceType:      "observer",
					LogAnalyzerType:    "ob_light",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			out := make(chan []*message.Message)
			toBeStopped := make(chan bool)
			logTailerExecutor := NewLogTailerExecutor(tt.args.tailConf, monagent.RecoveryConfig{}, toBeStopped, out)
			if err != nil {
				t.Fatal(err)
			}
			go func() {
				logTailerExecutor.fileProcessQueue.queue = tt.args.initLogQueue
				err := logTailerExecutor.handleFileQueue(context.Background())
				if err != nil {
					t.Error(err)
				}
				close(out)
			}()

			for alarmMessages := range out {
				if len(alarmMessages) > 0 {
					atomic.StoreInt64(&invocationCnt, atomic.LoadInt64(&invocationCnt)+int64(len(alarmMessages)))
				}
				cnt := atomic.LoadInt64(&invocationCnt)
				if cnt == int64(totalCnt) {
					close(toBeStopped)
				}
			}
		})
	}
}

func consumeOutputByCount(out chan []*message.Message, count int) {
	i := 0
	for range out {
		if i++; i >= count {
			break
		}
	}
}

func TestLogTailer_Run(t *testing.T) {
	invocationCnt := int64(0)
	logEntryCnt, extraLogEntryCnt := 1, 1

	tmpDir, err := prepareTestDirTree("tmp1")
	if err != nil {
		return
	}
	defer os.RemoveAll(tmpDir)
	logFile := "ob.log.wf"

	tmpFile, err := createAndWriteFile(tmpDir, logFile, 0, logEntryCnt)
	defer tmpFile.Close()
	logTailer, err := NewLogTailer(
		monagent.LogTailerConfig{
			TailConfigs: []monagent.TailConfig{
				{
					LogDir:             tmpDir,
					LogFileName:        logFile,
					ProcessLogInterval: time.Millisecond,
					LogAnalyzerType:    "ob_light",
				},
			},
			RecoveryConfig: monagent.RecoveryConfig{},
		})
	if err != nil {
		t.Fatal(err)
	}

	out := make(chan []*message.Message)
	err = logTailer.Run(context.Background(), out)
	if err != nil {
		t.Fatal(err)
	}

	consumeOutputByCount(out, logEntryCnt+extraLogEntryCnt)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "监听日志目录，中间更新了文件",
		},
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			now := time.Now()
			err = os.Rename(tmpFile.Name(), fmt.Sprintf("%s.%s", tmpFile.Name(), now.Format(logTimeInFileNameLayout)))
			if err != nil {
				t.Fatal(err)
			}
			tmpFile1, err := createAndWriteFile(tmpDir, logFile, logEntryCnt, 2*logEntryCnt)
			if err != nil {
				t.Fatal()
			}
			defer tmpFile1.Close()

			err = os.Rename(tmpFile1.Name(), fmt.Sprintf("%s.%s", tmpFile1.Name(),
				now.Add(time.Second).Format(logTimeInFileNameLayout)))
			if err != nil {
				t.Fatal(err)
			}

			tmpFile2, err := createAndWriteFile(tmpDir, logFile, 2*logEntryCnt, 3*logEntryCnt)
			if err != nil {
				t.Fatal()
			}
			defer tmpFile2.Close()

			for alarmMessages := range out {
				if len(alarmMessages) > 0 {
					invocationCnt += int64(len(alarmMessages))
				}
				if invocationCnt == int64(2*(logEntryCnt+extraLogEntryCnt)) {
					logTailer.Stop()
					close(out)
				}
			}
		})
	}
}
