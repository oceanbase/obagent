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
