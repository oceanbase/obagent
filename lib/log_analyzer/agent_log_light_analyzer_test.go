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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentLogLightAnalyzer_ParseLine(t *testing.T) {
	rawLogLine := `2022-03-23T15:46:34.78666+08:00 INFO [115773,] caller=shell/exec.go:87:execute: execute shell command start, command=Command{user=root, program=sh, cmd=netstat -tunlp 2>/dev/null | { grep '115772/' || true; }, timeout=10s} fields: duration="229.172µs"`
	logAnalyzer := NewAgentLogLightAnalyzer("mgragent.log")
	msg, isNewLine := logAnalyzer.ParseLine(rawLogLine)
	assert.Equal(t, true, isNewLine)
	checkTag(msg, "level", "INFO", t)
	assert.Equal(t, "mgragent.log", msg.GetName())
	expectedLogAt, _ := time.Parse(agentLogTimeFormat, "2022-03-23T15:46:34.78666+08:00")
	assert.Equal(t, expectedLogAt, msg.GetTime())
	raw, ok := msg.GetField("raw")
	assert.Equal(t, true, ok)
	rawStr := raw.(string)
	assert.Equal(t, rawLogLine, rawStr)
}
