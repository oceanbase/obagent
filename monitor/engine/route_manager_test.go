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

package engine

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRouteManagerPipelineFromPipelineGroup(t *testing.T) {
	testPipelineInstance := &Pipeline{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineInstance)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}

	tests := []struct {
		name   string
		fields *RouteManager
	}{
		{name: "test", fields: GetRouteManager()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := tt.fields

			manager.AddPipelineGroup("/test", testPipelineInstance)

			pipelineGroup, _ := manager.GetPipelineGroup("/test")
			pipelineInstance := pipelineGroup.Front().Value.(*Pipeline)

			manager.DeletePipelineGroup("/test", testPipelineInstance)

			if deepEqual := reflect.DeepEqual(pipelineInstance, testPipelineInstance); !deepEqual {
				t.Errorf("route manager test deep equal pipeline instance failed %s", err.Error())
			}

		})
	}
}

var testJSONModule = `
{
	"module": "test",
	"testInput": {
		"plugin": "test",
		"config": {
			"timeout": 10,
			"pluginConfig": null
		}
	},
	"testProcessor": {
		"plugin": "test",
		"config": {
			"timeout": 10,
			"pluginConfig": null
		}
	},
	"testOutput": {
		"plugin": "test",
		"config": {
			"timeout": 10,
			"pluginConfig": null
		}
	},
	"testExporter": {
		"plugin": "test",
		"config": {
			"timeout": 10,
			"pluginConfig": null
		}
	},
	"pipelines": [{
			"name": "pipeline1",
			"config": {
				"scheduleStrategy": "trigger",
				"exposeUrl": "/metrics/test"
			},
			"structure": {
				"inputs": [{
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}],
				"processors": [{
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}],
				"exporter": {
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}
			}
		},
		{
			"name": "pipeline2",
			"config": {
				"scheduleStrategy": "periodic",
				"period": 5
			},
			"structure": {
				"inputs": [{
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}],
				"processors": [{
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}],
				"output": {
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}
			}
		}
	]
}
`
