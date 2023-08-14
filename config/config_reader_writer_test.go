/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeConfig(t *testing.T) {
	configContent := `configVersion: ""
configs:
    - key: string.not.empty
      value: foo
      valueType: string
    - key: string.empty
      value: 
      valueType: string
    - key: bool.true
      value: true
      valueType: bool
    - key: bool.empty
      value: 
      valueType: bool
    - key: bool.false
      value: false
      valueType: bool
    - key: bool.false.default-true
      value: false
      valueType: bool
    - key: int64.not.zore
      value: 200
      valueType: int64
    - key: int64.empty
      value: 
      valueType: int64
    - key: int64.zore
      value: 0
      valueType: int64`
	configPropertyMetas["string.not.empty"] = &ConfigProperty{
		DefaultValue: "bar",
		ValueType:    ValueString,
	}
	configPropertyMetas["string.empty"] = &ConfigProperty{
		DefaultValue: "bar for empty",
		ValueType:    ValueString,
	}
	configPropertyMetas["bool.true"] = &ConfigProperty{
		DefaultValue: "false",
		ValueType:    ValueBool,
	}
	configPropertyMetas["bool.empty"] = &ConfigProperty{
		DefaultValue: "false",
		ValueType:    ValueBool,
	}
	configPropertyMetas["bool.false"] = &ConfigProperty{
		DefaultValue: "false",
		ValueType:    ValueBool,
	}
	configPropertyMetas["bool.false.default-true"] = &ConfigProperty{
		DefaultValue: "true",
		ValueType:    ValueBool,
	}
	configPropertyMetas["int64.not.zore"] = &ConfigProperty{
		DefaultValue: "0",
		ValueType:    ValueInt64,
	}
	configPropertyMetas["int64.empty"] = &ConfigProperty{
		DefaultValue: "1",
		ValueType:    ValueInt64,
	}
	configPropertyMetas["int64.zore"] = &ConfigProperty{
		DefaultValue: "1",
		ValueType:    ValueInt64,
	}
	asserts := map[string]interface{}{
		"string.not.empty":        "foo",
		"string.empty":            "",
		"bool.true":               true,
		"bool.empty":              false,
		"bool.false":              false,
		"bool.false.default-true": false,
		"int64.not.zore":          int64(200),
		"int64.empty":             int64(0),
		"int64.zore":              int64(0),
	}
	group, err := decodeConfigPropertiesGroup(context.Background(), []byte(configContent))
	assert.Nil(t, err)
	configs := group.Configs
	for _, config := range configs {
		ass, ex := asserts[config.Key]
		if !ex {
			continue
		}
		assert.Equalf(t, ass, config.Val(), "key %s", config.Key)
		assert.Equalf(t, ass, config.Value, "key %s", config.Key)
	}
}
