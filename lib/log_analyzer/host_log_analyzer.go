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

// Nov 28 03:20:05
const hostLogAtLayout = "Jan _2 15:04:05"
const logTimeInHostFileNameLayout = "20060102"

// HostLogAnalyzer 主机日志
type HostLogAnalyzer struct {
	fileName string
}

var hostLogRegexp = regexp.MustCompile(`^(?P<time>[A-Z][a-z]{2,} [ |\d]\d \d\d:\d\d:\d\d) (?P<host>[^ ]+) (?P<content>.+)`)
var hostSubMatchIndex = make(map[string]int)

func init() {
	for i, name := range hostLogRegexp.SubexpNames() {
		hostSubMatchIndex[name] = i
	}
}

func NewHostLogAnalyzer(fileName string) LogAnalyzer {
	return &HostLogAnalyzer{
		fileName: fileName,
	}
}

func (o *HostLogAnalyzer) ParseLine(line string) (*message.Message, bool) {
	subMatch := hostLogRegexp.FindStringSubmatch(line)
	if subMatch == nil {
		return nil, false
	}
	contentIdx := hostSubMatchIndex["content"]
	if contentIdx >= len(subMatch) {
		return nil, false
	}
	t, err := parseHostLogTime(subMatch[hostSubMatchIndex["time"]], time.Now())
	if err != nil {
		return nil, false
	}
	msg := message.NewMessage(o.fileName, message.Log, t)
	msg.AddField("raw", line)
	msg.AddField("content", subMatch[contentIdx])
	msg.AddTag("level", "info")
	return msg, true
}

func parseHostLogTime(logStr string, now time.Time) (time.Time, error) {
	logAt, err := time.ParseInLocation(hostLogAtLayout, logStr, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	// /var/log 日志默认保存 4 周
	// 由于日志行中没有年份，这里特殊处理下年份，对 12 月做特殊处理
	year := now.Year()
	if logAt.Month() == time.December &&
		now.Month() == time.January &&
		logAt.After(now) {
		year -= 1
	}
	logAt = time.Date(year, logAt.Month(), logAt.Day(), logAt.Hour(), logAt.Minute(), logAt.Second(), logAt.Nanosecond(), time.Local)
	return logAt, nil
}

// GetFileEndTime 获取文件最后写入时间
func (o *HostLogAnalyzer) GetFileEndTime(info os.FileInfo) (time.Time, error) {
	return ParseTimeFromFileName(info.Name(), "-", logTimeInHostFileNameLayout, info.ModTime()), nil
}
