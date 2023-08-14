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
	"reflect"
	"testing"

	"github.com/oceanbase/obagent/errors"
)

func TestConfigProperty_Parse(t *testing.T) {
	tests := []struct {
		name     string
		property *ConfigProperty
		wantVal  interface{}
		wantErr  bool
	}{
		// parse string
		{
			name: "parse normal string",
			property: &ConfigProperty{
				Key:       "key1",
				Value:     "value1",
				ValueType: ValueString,
			},
			wantVal: "value1",
			wantErr: false,
		},
		{
			name: "parse nil string",
			property: &ConfigProperty{
				Key:       "key2",
				Value:     "",
				ValueType: ValueString,
			},
			wantVal: "",
			wantErr: false,
		},

		// parse bool
		{
			name: "parse true",
			property: &ConfigProperty{
				Key:       "key1",
				Value:     true,
				ValueType: ValueBool,
			},
			wantVal: true,
			wantErr: false,
		},
		{
			name: "parse nil string",
			property: &ConfigProperty{
				Key:       "key2",
				Value:     false,
				ValueType: ValueBool,
			},
			wantVal: false,
			wantErr: false,
		},
		{
			name: "parse nil bool",
			property: &ConfigProperty{
				Key:       "key3",
				Value:     nil,
				ValueType: ValueBool,
			},
			wantVal: false,
			wantErr: false,
		},

		// parse int
		{
			name: "parse positive int64",
			property: &ConfigProperty{
				Key:       "key1",
				Value:     100,
				ValueType: ValueInt64,
			},
			wantVal: int64(100),
			wantErr: false,
		},
		{
			name: "parse negative int64",
			property: &ConfigProperty{
				Key:       "key2",
				Value:     -100,
				ValueType: ValueInt64,
			},
			wantVal: int64(-100),
			wantErr: false,
		},
		{
			name: "parse nil int64",
			property: &ConfigProperty{
				Key:       "key3",
				Value:     nil,
				ValueType: ValueInt64,
			},
			wantVal: int64(0),
			wantErr: false,
		},

		// parse float
		{
			name: "parse positive float",
			property: &ConfigProperty{
				Key:       "key1",
				Value:     100.0,
				ValueType: ValueFloat64,
			},
			wantVal: 100.0,
			wantErr: false,
		},
		{
			name: "parse negative float",
			property: &ConfigProperty{
				Key:       "key2",
				Value:     -100.0,
				ValueType: ValueFloat64,
			},
			wantVal: -100.0,
			wantErr: false,
		},
		{
			name: "parse nil float64",
			property: &ConfigProperty{
				Key:       "key3",
				Value:     nil,
				ValueType: ValueFloat64,
			},
			wantVal: nil,
			wantErr: true,
		},

		{
			name: "parse none valueType",
			property: &ConfigProperty{
				Key:       "key3",
				Value:     "nil",
				ValueType: ValueType("none"),
			},
			wantVal: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := tt.property.Parse(tt.property.Value)
			if (err != nil) != tt.wantErr {
				t.Errorf("name %s, ConfigProperty.Parse() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("name %s, ConfigProperty.Parse() = %v, want %v", tt.name, gotVal, tt.wantVal)
			}
		})
	}
}

func TestConfigProperty_ParseValid(t *testing.T) {
	tests := []struct {
		name     string
		property *ConfigProperty
		wantVal  interface{}
		wantErr  bool
	}{
		// parse string
		{
			name: "parse normal string",
			property: &ConfigProperty{
				Key:       "key1",
				Value:     "value1",
				ValueType: ValueString,
				Valid:     func(value interface{}) error { return nil },
			},
			wantVal: "value1",
			wantErr: false,
		},

		// parse int
		{
			name: "parse positive",
			property: &ConfigProperty{
				Key:       "key2",
				Value:     100,
				ValueType: ValueInt64,
				Valid: func(value interface{}) error {
					if val := value.(int64); val > 10 {
						return errors.Errorf("keys should be less than 11")
					}
					return nil
				},
			},
			wantVal: int64(100),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := tt.property.Parse(tt.property.Value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigProperty.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("ConfigProperty.Parse() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}
