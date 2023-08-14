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

package log_tailer

import (
	"github.com/oceanbase/obagent/monitor/plugins/common"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/lib/log_analyzer"
)

func Test_logAnalyzer_isErrLog(t *testing.T) {
	Convey("日志长度 < 34 的情况", t, func() {
		logAnalyzer := log_analyzer.NewObLogLightAnalyzer("test.log")
		logLineInfo, isNewLine := logAnalyzer.ParseLine("test log")
		So(isNewLine, ShouldBeFalse)
		logLevel, _ := logLineInfo.GetTag(common.Level)
		So(logLevel == "ERROR", ShouldEqual, false)
	})
}
