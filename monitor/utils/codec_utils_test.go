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

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeYaml(t *testing.T) {
	yamlStr := `
      k1: v1
      k2: [1,2,3]
    `
	config, err := DecodeYaml(yamlStr)
	require.True(t, err == nil)
	require.Equal(t, "v1", config["k1"])
	require.Equal(t, []interface{}{1, 2, 3}, config["k2"])
}

func TestDecodeYamlFail(t *testing.T) {
	yamlStr := `
      k1: v1
      k2: [1,2,3],123
    `
	_, err := DecodeYaml(yamlStr)
	require.True(t, err != nil)
}

func TestEncodeYaml(t *testing.T) {
	config := make(map[string]interface{})
	_, err := EncodeYaml(config)
	require.True(t, err == nil)
}
