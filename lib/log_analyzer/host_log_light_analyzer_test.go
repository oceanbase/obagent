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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHostLogLightAnalyzer_ParseLine(t *testing.T) {
	rawLogLine := `Mar 28 03:37:07 h07b11215.sqa.eu95 run-parts(/etc/cron.daily)[92918]: starting rpm`
	logAnalyzer := NewHostLogLightAnalyzer("messages")
	msg, isNewLine := logAnalyzer.ParseLine(rawLogLine)
	assert.Equal(t, true, isNewLine)
	assert.Equal(t, "messages", msg.GetName())

	expectedLogAt := time.Date(2023, 3, 28, 3, 37, 7, 0, time.Local)
	assert.Equal(t, expectedLogAt, msg.GetTime())

	raw, ok := msg.GetField("raw")
	assert.Equal(t, true, ok)
	contentStr := raw.(string)
	assert.Equal(t, rawLogLine, contentStr)
}
