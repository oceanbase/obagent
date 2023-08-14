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

package log_query

import (
	"os"
	"time"

	"github.com/oceanbase/obagent/lib/log_analyzer"
)

type ConditionType string

const (
	text  ConditionType = "TEXT"
	regex ConditionType = "REGEX"
)

type QueryLogRequest struct {
	StartTime           time.Time     `json:"startTime"`
	EndTime             time.Time     `json:"endTime"`
	LogType             string        `json:"logType"`
	Keyword             []string      `json:"keyword"`
	KeywordType         ConditionType `json:"keywordType"`
	ExcludeKeyword      []string      `json:"excludeKeyword"`
	ExcludeKeywordType  ConditionType `json:"excludeKeywordType"`
	LogLevel            []string      `json:"logLevel"`
	ReqId               string        `json:"reqId"`
	LastQueryFileId     uint64        `json:"lastQueryFileId"`
	LastQueryFileOffset int64         `json:"lastQueryFileOffset"`
	Limit               int64         `json:"limit"`
}

type DirAndFilePattern struct {
	LogAnalyzerCategory string
	Dir                 string
	LogFilePatterns     []string
}

type FileDetailInfo struct {
	LogAnalyzer log_analyzer.LogAnalyzer
	Dir         string
	FileInfo    os.FileInfo
	FileTime    time.Time
	FileDesc    *os.File
	FileId      uint64
	FileOffset  int64
}

type FileInfo struct {
	FileName   string
	FileId     uint64
	FileOffset int64
}

type LogEntry struct {
	LogAt                       time.Time `json:"logAt"`
	LogLine                     []byte    `json:"logLine"`
	LogLevel                    string    `json:"logLevel"`
	FileName                    string    `json:"fileName"`
	FileId                      uint64    `json:"fileId"`
	FileOffset                  int64     `json:"fileOffset"`
	isMatchedByLogAtAndLogLevel bool
	isMatched                   bool
}

type Position struct {
	FileId     uint64 `json:"fileId"`
	FileOffset int64  `json:"fileOffset"`
}

func (q *QueryLogRequest) validate() bool {
	if q.StartTime.IsZero() ||
		q.EndTime.IsZero() ||
		q.LogType == "" {
		return false
	}
	return true
}
