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

func TestObLogLightAnalyzer_ParseLine(t *testing.T) {
	rawLogLine := `[2022-01-20 10:49:14.332262] INFO  [LIB] ob_json.cpp:278 [3451815][274][Y0-0000000000000000] [lt=14] [dc=0] invalid token type, maybe it is valid empty json type(cur_token_.type=93, ret=-5006)`
	logAnalyzer := NewObLogLightAnalyzer("observer.log")
	msg, isNewLine := logAnalyzer.ParseLine(rawLogLine)
	assert.Equal(t, true, isNewLine)
	assert.Equal(t, "observer.log", msg.GetName())

	expectedLogAt, err := time.ParseInLocation(observerLogAtLayout, "2022-01-20 10:49:14.332262", time.Local)
	assert.NoError(t, err)
	assert.Equal(t, expectedLogAt, msg.GetTime())

	checkTag(msg, "level", "INFO", t)

	raw, ok := msg.GetField("raw")
	assert.Equal(t, true, ok)
	contentStr := raw.(string)
	assert.Equal(t, rawLogLine, contentStr)

	errCode, ok := msg.GetField("errCode")
	assert.Equal(t, true, ok)
	parsedErrCode := errCode.(int)
	assert.Equal(t, 5006, parsedErrCode)
}
