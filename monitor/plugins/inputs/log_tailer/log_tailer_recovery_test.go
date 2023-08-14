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

package log_tailer

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/config/monagent"
)

func TestLogTailer_storeLastPosition(t *testing.T) {
	logDir, err := prepareTestDirTree("log_dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(logDir)

	logFileName := fmt.Sprintf("%s/%s", logDir, "ob.log.wf")
	tmpLogFile, err := os.Create(logFileName)
	if err != nil {
		t.Fatal(err)
	}
	tmpLogFileStat, err := tmpLogFile.Stat()
	if err != nil {
		t.Fatal(err)
	}
	tmpLogFileModTime := tmpLogFileStat.ModTime()
	if err != nil {
		t.Fatal(err)
	}

	tailConfig := monagent.TailConfig{
		LogDir:             logDir,
		LogFileName:        "ob.log.wf",
		ProcessLogInterval: 100,
		LogSourceType:      "observer",
		LogAnalyzerType:    "ob_light",
	}

	positionStoreStr := "position_store"
	positionStoreDir, err := prepareTestDirTree(positionStoreStr)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(positionStoreDir)
	recoveryConfig := monagent.RecoveryConfig{
		Enabled:              true,
		LastPositionStoreDir: positionStoreDir,
	}

	Convey("storeLastPosition", t, func() {
		toBeStoredFileInfo := &logFileInfo{
			logSourceType:   "observer",
			fileName:        "ob.log.wf",
			fileDesc:        tmpLogFile,
			fileOffset:      3333,
			offsetLineLogAt: tmpLogFileModTime,
		}
		err = storeLastPosition(context.Background(), recoveryConfig, toBeStoredFileInfo)
		if err != nil {
			t.Fatal(err)
		}

		loadedLogFileInfo, err := loadLastPosition(context.Background(), recoveryConfig, tailConfig)
		if err != nil {
			t.Fatal(err)
		}

		So(toBeStoredFileInfo.fileOffset, ShouldEqual, loadedLogFileInfo.fileOffset)
		So(toBeStoredFileInfo.fileName, ShouldEqual, loadedLogFileInfo.fileName)
	})
}
