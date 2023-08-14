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

	"github.com/oceanbase/obagent/config/monagent"
)

func TestCreateModule(t *testing.T) {
	testPipelineModule := &monagent.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test create json module failed %s", err.Error())
		return
	}
	testPipelineInstances, _ := CreatePipelines(testPipelineModule)

	type args struct {
		pipelineModule *monagent.PipelineModule
	}
	tests := []struct {
		name string
		args args
		want []*Pipeline
	}{
		{name: "test", args: args{testPipelineModule}, want: testPipelineInstances},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := CreatePipelines(tt.args.pipelineModule); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePipelineInstances() = %v, want %v", got, tt.want)
			}
		})
	}
}
