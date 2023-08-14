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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentLogAnalyzer_ParseLine(t *testing.T) {
	rawLogLine := `2022-03-23T15:46:34.78666+08:00 INFO [115773,] caller=shell/exec.go:87:execute: execute shell command start, command=Command{user=root, program=sh, cmd=netstat -tunlp 2>/dev/null | { grep '115772/' || true; }, timeout=10s} fields: duration="229.172µs"`
	logAnalyzer := NewAgentLogAnalyzer("mgragent.log")
	msg, isNewLine := logAnalyzer.ParseLine(rawLogLine)
	assert.Equal(t, true, isNewLine)
	checkTag(msg, "level", "info", t)
	checkTag(msg, "pid", "115773", t)
	checkTag(msg, "source", "shell/exec.go", t)
	content, ok := msg.GetField("content")
	assert.Equal(t, true, ok)
	contentStr := content.(string)
	assert.Equal(t, `execute shell command start, command=Command{user=root, program=sh, cmd=netstat -tunlp 2>/dev/null | { grep '115772/' || true; }, timeout=10s} fields: duration="229.172µs"`, contentStr)
}
