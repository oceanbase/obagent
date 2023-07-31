package agentd

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLimitConfigUnmarshal(t *testing.T) {
	content := `cpuQuota: 1.0
memoryQuota: 1024MB`
	var conf LimitConfig
	err := yaml.Unmarshal([]byte(content), &conf)
	assert.Nil(t, err)
	assert.Equal(t, int64(1024*1024*1024), int64(conf.MemoryQuota))

	logrus.Infof("conf: %+v", conf)
}

func TestEmptyQuotaLimitConfigUnmarshal(t *testing.T) {
	content := `cpuQuota: 
memoryQuota: `
	var conf LimitConfig
	err := yaml.Unmarshal([]byte(content), &conf)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), int64(conf.MemoryQuota))

	logrus.Infof("conf: %+v", conf)
}
