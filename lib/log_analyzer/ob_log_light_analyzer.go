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

package log_analyzer

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oceanbase/obagent/monitor/message"
)

const observerLogAtLayout = "2006-01-02 15:04:05.000000"
const logTimeInFileNameLayout = "20060102150405"

var (
	obLogAtRegexp    = regexp.MustCompile(`^\[\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d\.\d\d\d\d\d\d\]`)
	errCodeRegexp    = regexp.MustCompile(`ret=-\d+`)
	obLogLevelRegexp = regexp.MustCompile(defaultLogLevels)
)

type ObLogLightAnalyzer struct {
	fileName string
}

func NewObLogLightAnalyzer(fileName string) LogAnalyzer {
	return &ObLogLightAnalyzer{
		fileName: fileName,
	}
}

func (o *ObLogLightAnalyzer) ParseLine(logLine string) (*message.Message, bool) {
	logAt := time.Time{}
	isNewLine := false
	matchedLogAt := obLogAtRegexp.FindString(logLine)
	if matchedLogAt != "" {
		timeStr := strings.TrimRight(strings.TrimLeft(matchedLogAt, "["), "]")
		parsedLogAt, err := time.ParseInLocation(observerLogAtLayout, timeStr, time.Local)
		if err != nil {
			// 还是会构造 message
			// TODO 仅需 log
		} else {
			logAt = parsedLogAt
			isNewLine = true
		}
	}

	errCode, err := o.getErrCode(logLine)
	if err != nil {
		// TODO 仅需 log
	}

	msg := message.NewMessage(o.fileName, message.Log, logAt)
	msg.AddField("raw", logLine)
	msg.AddTag("level", obLogLevelRegexp.FindString(logLine))
	msg.AddField("errCode", errCode)

	return msg, isNewLine
}

// GetFileEndTime 获取文件最后写入时间 TODO 复用整合
func (o *ObLogLightAnalyzer) GetFileEndTime(info os.FileInfo) (time.Time, error) {
	fileTime := info.ModTime()
	fileName := info.Name()
	lastDotIdx := strings.LastIndex(fileName, ".")
	timeStr := fileName[lastDotIdx+1:]
	if lastDotIdx == -1 || timeStr == "" {
		return info.ModTime(), nil
	}
	parsedTime, err := time.ParseInLocation(logTimeInFileNameLayout, timeStr, time.Local)
	if err != nil {
		return fileTime, nil
	}
	return parsedTime, nil
}

// getErrCode 获取日志中的错误码
func (o *ObLogLightAnalyzer) getErrCode(logLine string) (int, error) {
	matchedErrCodes := errCodeRegexp.FindAllString(logLine, -1)
	matchedErrCodesLen := len(matchedErrCodes)
	if matchedErrCodesLen > 0 {
		lastErrCodeStr := matchedErrCodes[matchedErrCodesLen-1]
		// 匹配的格式为 ret=-\d+，数字从下标位置 5 开始
		if len(lastErrCodeStr) >= 5 {
			errCode, err := strconv.Atoi(lastErrCodeStr[5:])
			if err != nil {
				return -1, err
			}
			return errCode, nil
		}
	}

	return -1, nil
}
