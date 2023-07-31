package log_analyzer

import (
	"regexp"
	"strings"
)

const defaultLogLevels = "TRACE|DEBUG|INFO|WARN|ERROR|FATAL"

// GetLogType 获取日志类型 observer election rootservice obproxy monagent mgragent 等
// 如传入 election.log.wf 返回
func GetLogType(fileName string) string {
	idx := strings.Index(fileName, ".")
	if idx == -1 {
		return fileName
	}
	return fileName[:idx]
}

// subExpIndex go 1.14 中没有该实现，这里做个兼容，后面升级了版本后可以直接用官方 sdk
func subExpIndex(re *regexp.Regexp, name string) int {
	if name != "" {
		for i, s := range re.SubexpNames() {
			if name == s {
				return i
			}
		}
	}
	return -1
}
