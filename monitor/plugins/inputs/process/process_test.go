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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestProcessExists(t *testing.T) {
	config := `
        processNames: [process1]
    `
	var configMap map[string]interface{}
	yaml.Unmarshal([]byte(config), &configMap)
	processInput := &ProcessInput{}
	processInput.Init(context.Background(), configMap)
	allProcessNames = func() ([]string, error) {
		return []string{"process0", "process1", "process2"}, nil
	}
	metrics, err := processInput.CollectMsgs(context.Background())
	require.Equal(t, 1, len(metrics))
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
        processNames: [process1]
    `
	var configMap map[string]interface{}
	yaml.Unmarshal([]byte(config), &configMap)
	processInput := &ProcessInput{}
	processInput.Init(context.Background(), configMap)
	allProcessNames = func() ([]string, error) {
		return []string{"process0", "process2"}, nil
	}
	metrics, err := processInput.CollectMsgs(context.Background())
	require.Equal(t, 1, len(metrics))
	require.True(t, err == nil)
	require.Equal(t, 1, len(metrics))
	value, exists := metrics[0].GetField("value")
	require.True(t, exists)
	value, ok := value.(float64)
	require.True(t, ok)
	require.Equal(t, 0.0, value)
}

func TestErr(t *testing.T) {
	config := `
        processNames: [process1]
    `
	var configMap map[string]interface{}
	yaml.Unmarshal([]byte(config), &configMap)
	processInput := &ProcessInput{}
	processInput.Init(context.Background(), configMap)
	allProcessNames = func() ([]string, error) {
		return nil, errors.New("test")
	}
	metrics, err := processInput.CollectMsgs(context.Background())
	require.Equal(t, 0, len(metrics))
	require.True(t, err == nil)
}
