package log_analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogType(t *testing.T) {
	logType := GetLogType("observer.log.wf.20220116")
	assert.Equal(t, "observer", logType)
}
