package monagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineModuleStatus_Validate(t *testing.T) {
	var pms PipelineModuleStatus = "false"
	assert.False(t, pms.Validate())
	pms = "active"
	assert.True(t, pms.Validate())
}
