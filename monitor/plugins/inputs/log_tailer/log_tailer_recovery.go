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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
)

func getLastPositionStorePath(lastPositionStoreDir, logSourceType, logFileName string) string {
	return filepath.Join(lastPositionStoreDir, logSourceType+"_"+logFileName+".txt")
}

// storeLastPosition Persist the last query location
func storeLastPosition(ctx context.Context, recoveryConf monagent.RecoveryConfig, fileInfo *logFileInfo) error {
	if !recoveryConf.Enabled || fileInfo == nil {
		return nil
	}
	positionStoreFilePath := getLastPositionStorePath(recoveryConf.LastPositionStoreDir, fileInfo.logSourceType, fileInfo.fileName)
	f, err := os.OpenFile(positionStoreFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	logFileStatInfo, err := libFile.GetFileStatInfo(ctx, fileInfo.fileDesc)
	if err != nil {
		return err
	}
	recoveryInfo := &RecoveryInfo{
		FileName:   fileInfo.fileName,
		FileId:     logFileStatInfo.FileId(),
		DevId:      logFileStatInfo.DevId(),
		FileOffset: fileInfo.fileOffset,
		TimePoint:  fileInfo.offsetLineLogAt,
	}

	err = storeRecoveryInfoToWriter(ctx, recoveryInfo, f)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

func storeRecoveryInfoToWriter(ctx context.Context, recoveryInfo *RecoveryInfo, f io.Writer) error {
	recoveryInfoBytes, err := json.Marshal(recoveryInfo)
	if err != nil {
		return err
	}
	_, err = f.Write(recoveryInfoBytes)
	if err != nil {
		return err
	}

	return nil
}

// loadLastPosition Load the last queried location
// If the last traced file cannot be found (it may have been cleaned up),
// the earliest (first) file found within the time of recoveryInfo.TimePoint ~ now is returned
func loadLastPosition(ctx context.Context, recoveryConf monagent.RecoveryConfig, conf monagent.TailConfig) (*logFileInfo, error) {
	lastPositionStoreFile := getLastPositionStorePath(recoveryConf.LastPositionStoreDir, conf.LogSourceType, conf.LogFileName)
	ctxLog := log.WithContext(ctx).WithField("lastPositionStoreFile", lastPositionStoreFile)
	f, err := checkAndOpenFile(ctx, lastPositionStoreFile)
	if err != nil {
		return nil, err
	}
	defer closeFile(ctx, f)
	recoveryInfo, err := loadPositionFromReader(f)
	if err != nil {
		return nil, err
	}
	if recoveryInfo == nil {
		return nil, nil
	}

	start, end := recoveryInfo.TimePoint, time.Now()
	newLogFiles, err := getLogsWithinTime(ctx, conf, start, end)
	if err != nil {
		ctxLog.WithError(err).WithFields(log.Fields{"start": start, "end": end}).Warn("check isSameFile failed")
		return nil, err
	}
	if len(newLogFiles) == 0 {
		return nil, nil
	}

	firstNewLogFile := newLogFiles[0]
	newLogFiles = newLogFiles[1:]

	filterFunc := func(ctx context.Context, fd *os.File) (isFilter bool) {
		ctxLog1 := log.WithContext(ctx)
		fdStat, err1 := libFile.GetFileStatInfo(ctx, fd)
		if err1 != nil {
			ctxLog1.WithError(err1).Warn("GetFileStatInfo failed")
			return true
		}
		return !(recoveryInfo.FileId == fdStat.FileId() && recoveryInfo.DevId == fdStat.DevId())
	}
	matchedFiles, unmatchedFiles, err := logFileFilterIn(ctx, newLogFiles, filterFunc)
	closeFiles(ctx, unmatchedFiles)
	if err != nil {
		closeFiles(ctx, matchedFiles)
		return nil, err
	}

	ret := &logFileInfo{
		logSourceType:   conf.LogSourceType,
		logAnalyzerType: conf.LogAnalyzerType,
		fileName:        conf.LogFileName,
		fileDesc:        firstNewLogFile,
	}

	if len(matchedFiles) != 0 {
		closeFile(ctx, firstNewLogFile)
		ret.fileDesc = matchedFiles[0]
	}
	firstNewLogStat, err := libFile.GetFileStatInfo(ctx, firstNewLogFile)
	if err != nil {
		return nil, err
	}
	if len(matchedFiles) != 0 || (firstNewLogStat.FileId() == recoveryInfo.FileId && firstNewLogStat.DevId() == recoveryInfo.DevId) {
		_, err1 := ret.fileDesc.Seek(recoveryInfo.FileOffset, 0)
		if err1 != nil {
			return nil, err1
		}
		ret.fileOffset = recoveryInfo.FileOffset
		ret.offsetLineLogAt = recoveryInfo.TimePoint
	}

	return ret, nil
}

func loadPositionFromReader(f io.Reader) (*RecoveryInfo, error) {
	reader := bufio.NewReader(f)
	var b bytes.Buffer
	_, err := io.Copy(&b, reader)
	if err != nil {
		return nil, err
	}

	if len(b.Bytes()) == 0 {
		return nil, nil
	}

	recoveryInfo := &RecoveryInfo{}

	err = json.Unmarshal(b.Bytes(), recoveryInfo)
	if err != nil {
		return nil, err
	}

	return recoveryInfo, nil
}
