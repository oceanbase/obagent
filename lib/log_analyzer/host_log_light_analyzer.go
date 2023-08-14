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
	"time"

	"github.com/oceanbase/obagent/monitor/message"
)

var (
	logAtRegexp = regexp.MustCompile(`^[A-Z][a-z]{2,} [ |\d]\d \d\d:\d\d:\d\d`)
)

// HostLogAnalyzer 主机日志
type HostLogLightAnalyzer struct {
	fileName string
}

func NewHostLogLightAnalyzer(fileName string) LogAnalyzer {
	return &HostLogLightAnalyzer{
		fileName: fileName,
	}
}

func (o *HostLogLightAnalyzer) ParseLine(line string) (*message.Message, bool) {
	logAt := time.Time{}
	isNewLine := false
	matchedLogAt := logAtRegexp.FindString(line)
	if matchedLogAt != "" {
		parsedLogAt, err := parseHostLogTime(matchedLogAt, time.Now().Local())
		if err != nil {
			// 还是会构造 message
			// TODO 仅需 log
		} else {
			logAt = parsedLogAt
			isNewLine = true
		}
	}

	msg := message.NewMessage(o.fileName, message.Log, logAt)
	msg.AddField("raw", line)
	return msg, isNewLine
}

// GetFileEndTime 获取文件最后写入时间
func (o *HostLogLightAnalyzer) GetFileEndTime(info os.FileInfo) (time.Time, error) {
	return ParseTimeFromFileName(info.Name(), "-", logTimeInHostFileNameLayout, info.ModTime()), nil
}
