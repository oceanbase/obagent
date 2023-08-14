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
	"os"
	"time"
)

type logFileInfo struct {
	logSourceType   string
	logAnalyzerType string
	fileName        string
	fileDesc        *os.File
	fileOffset      int64
	offsetLineLogAt time.Time
	isRenamed       bool
}

// RecoveryInfo Persistence information to record the query location when closed
type RecoveryInfo struct {
	FileName   string    `json:"fileName" yaml:"fileName"`
	FileId     uint64    `json:"fileId" yaml:"fileId"`
	DevId      uint64    `json:"devId" yaml:"devId"`
	FileOffset int64     `json:"fileOffset" yaml:"fileOffset"`
	TimePoint  time.Time `json:"timePoint" yaml:"timePoint"`
}

func (r RecoveryInfo) GetFileId() uint64 {
	return r.FileId
}

func (r RecoveryInfo) GetDevId() uint64 {
	return r.DevId
}
