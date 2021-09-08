// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package engine

import (
	"encoding/json"
	"testing"

	"github.com/oceanbase/obagent/config"
)

func TestPipelineInstance(t *testing.T) {
	testPipelineNode := &config.PipelineNode{}
	err := json.Unmarshal([]byte(testPipelinePullJSON), testPipelineNode)
	if err != nil {
		t.Errorf("test pipeline instance failed %s", err.Error())
	}
	pipelineInstance := createPipelineInstance(testPipelineNode)
	err = pipelineInstance.Init()
	if err != nil {
		t.Errorf("test pipeline instance init failed %s", err.Error())
	}
	pipelineInstance.Start()
	pipelineInstance.pipelinePush()
	pipelineInstance.Stop()

	testPipelineNode = &config.PipelineNode{}
	err = json.Unmarshal([]byte(testPipelinePushJSON), testPipelineNode)
	if err != nil {
		t.Errorf("test pipeline instance failed %s", err.Error())
	}
	pipelineInstance = createPipelineInstance(testPipelineNode)
	err = pipelineInstance.Init()
	if err != nil {
		t.Errorf("test pipeline instance init failed %s", err.Error())
	}
	pipelineInstance.Start()
	pipelineInstance.pipelinePush()
	pipelineInstance.Stop()
}

var testPipelinePullJSON = `
{
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
		"output": {
			"plugin": "test",
			"config": {
				"timeout": 10,
				"pluginConfig": null
			}
		}
	}
}
`
var testPipelinePushJSON = `
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
`
