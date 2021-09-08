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
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/config"
)

func TestRouteManagerPipelineFromPipelineGroup(t *testing.T) {
	testPipelineInstance := &PipelineInstance{}
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

			manager.addPipelineFromPipelineGroup("/test", testPipelineInstance)

			pipelineGroup, _ := manager.getPipelineGroup("/test")
			pipelineInstance := pipelineGroup.Front().Value.(*PipelineInstance)

			manager.delPipelineFromPipelineGroup("/test", testPipelineInstance)

			if deepEqual := reflect.DeepEqual(pipelineInstance, testPipelineInstance); !deepEqual {
				t.Errorf("route manager test deep equal pipeline instance failed %s", err.Error())
			}

		})
	}
}

func TestRoutePull(t *testing.T) {
	var testPipelineYaml = `name: pipeline
config:
 scheduleStrategy: trigger
 exposeUrl: /metrics/test
structure:
 inputs:
   - plugin: test
     config:
       timeout: 10s
       pluginConfig: {}
 processors:
   - plugin: test
     config:
       timeout: 10s
       pluginConfig: {}
 exporter:
   plugin: test
   config:
     timeout: 10s
     pluginConfig: {}`
	testPipelineNode := &config.PipelineNode{}
	err := yaml.Unmarshal([]byte(testPipelineYaml), testPipelineNode)
	if err != nil {
		t.Errorf("test pipeline instance failed %s", err.Error())
	}

	pipelineInstance := createPipelineInstance(testPipelineNode)
	err = pipelineInstance.pipeline.init()
	if err != nil {
		t.Errorf("test pipeline instance failed %s", err.Error())
		return
	}

	GetRouteManager().addPipelineFromPipelineGroup("/test", pipelineInstance)

	rt := &httpRoute{
		routePath: "/test",
	}
	responseRecorder := httptest.NewRecorder()
	r := &http.Request{}
	rt.ServeHTTP(responseRecorder, r)
}
