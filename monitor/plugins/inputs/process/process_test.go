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

package process

import (
	"context"
	"testing"

	"github.com/oceanbase/obagent/monitor/plugins/common"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestProcessExists(t *testing.T) {
	config := `
    collectConfig:
      - username: root
        processes: [process0]
    collect_interval: 1s
  `
	allProcess = func() []*common.ProcessInfo {
		return []*common.ProcessInfo{
			{
				Name:     "process0",
				Pid:      100,
				UserName: "root",
			},
		}
	}
	var configMap map[string]interface{}
	err := yaml.Unmarshal([]byte(config), &configMap)
	require.NoError(t, err)
	processInput := &ProcessInput{}
	err = processInput.Init(context.Background(), configMap)
	require.NoError(t, err)
	metrics, err := processInput.CollectMsgs(context.Background())
	require.True(t, err == nil)
	require.Equal(t, 1, len(metrics))
	value, exists := metrics[0].GetField("value")
	require.True(t, exists)
	value, ok := value.(float64)
	require.True(t, ok)
	require.Equal(t, 1.0, value)
}

func TestProcessNotExists(t *testing.T) {
	config := `
    collectConfig:
      - username: root
        processes: [process0]
    collect_interval: 1s
  `
	allProcess = func() []*common.ProcessInfo {
		return []*common.ProcessInfo{
			{
				Name:     "process1",
				Pid:      101,
				UserName: "root",
			},
		}
	}
	var configMap map[string]interface{}
	err := yaml.Unmarshal([]byte(config), &configMap)
	require.NoError(t, err)
	processInput := &ProcessInput{}
	err = processInput.Init(context.Background(), configMap)
	require.NoError(t, err)
	metrics, err := processInput.CollectMsgs(context.Background())
	require.True(t, err == nil)
	require.Equal(t, 1, len(metrics))
	value, exists := metrics[0].GetField("value")
	require.True(t, exists)
	value, ok := value.(float64)
	require.True(t, ok)
	require.Equal(t, 0.0, value)
}
