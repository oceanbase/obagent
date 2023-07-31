package shellf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCommandGroupMap(t *testing.T) {
	config, err := decodeConfig(configString)
	if err != nil {
		t.Error(err)
	}
	groupMap, err := buildCommandGroupMap(config.CommandGroups)
	assert.Nil(t, err)
	assert.NotEmpty(t, groupMap)
}
