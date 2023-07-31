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
