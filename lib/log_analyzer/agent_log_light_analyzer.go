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

package log_analyzer

import (
	"os"
	"regexp"
	"time"

	"github.com/oceanbase/obagent/monitor/message"
)

var (
	agentLogAtRegexp    = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d[.|\d]*[-|+]\d\d:\d\d`)
	agentLogLevelRegexp = regexp.MustCompile(defaultLogLevels)
)

type AgentLogLightAnalyzer struct {
	fileName string
}

func NewAgentLogLightAnalyzer(fileName string) LogAnalyzer {
	return &AgentLogLightAnalyzer{
		fileName: fileName,
	}
}

func (a *AgentLogLightAnalyzer) ParseLine(line string) (*message.Message, bool) {
	// 行首匹配的不是时间，那么说明不是新日志行，isNewLine 默认为 false
	logAt := time.Time{}
	isNewLine := false
	matchedLogAt := agentLogAtRegexp.FindString(line)
	if matchedLogAt != "" {
		parsedLogAt, err := time.ParseInLocation(agentLogTimeFormat, matchedLogAt, time.Local)
		if err != nil {
			// 还是会构造 message
			// TODO 仅需 log
		} else {
			logAt = parsedLogAt
			isNewLine = true
		}
	}

	msg := message.NewMessage(a.fileName, message.Log, logAt)
	msg.AddField("raw", line)
	msg.AddTag("level", agentLogLevelRegexp.FindString(line))

	return msg, isNewLine
}

// GetFileEndTime 获取文件最后写入时间
func (o *AgentLogLightAnalyzer) GetFileEndTime(info os.FileInfo) (time.Time, error) {
	return info.ModTime(), nil
}
