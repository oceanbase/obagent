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
