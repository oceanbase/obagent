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
