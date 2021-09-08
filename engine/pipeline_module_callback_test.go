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
	"testing"

	"github.com/oceanbase/obagent/config"
)

func TestAddPipeline(t *testing.T) {

	GetConfigManager().eventCallbackChan <- &configCallbackEvent{
		eventType:   addConfigEvent,
		execStatus:  configEventExecSucceed,
		description: "",
	}
	testPipelineModule := &config.PipelineModule{}
	err := addPipeline(testPipelineModule)
	if err != nil {
		t.Errorf("test add pipeline failed %s", err.Error())
	}

	GetConfigManager().eventCallbackChan <- &configCallbackEvent{
		eventType:   addConfigEvent,
		execStatus:  configEventExecFailed,
		description: "",
	}
	_ = addPipeline(testPipelineModule)

}

func TestDeletePipeline(t *testing.T) {

	GetConfigManager().eventCallbackChan <- &configCallbackEvent{
		eventType:   deleteConfigEvent,
		execStatus:  configEventExecSucceed,
		description: "",
	}
	testPipelineModule := &config.PipelineModule{}
	err := deletePipeline(testPipelineModule)
	if err != nil {
		t.Errorf("test delete pipeline failed %s", err.Error())
	}

	GetConfigManager().eventCallbackChan <- &configCallbackEvent{
		eventType:   deleteConfigEvent,
		execStatus:  configEventExecFailed,
		description: "",
	}
	_ = deletePipeline(testPipelineModule)

}

func TestUpdatePipeline(t *testing.T) {

	GetConfigManager().eventCallbackChan <- &configCallbackEvent{
		eventType:   updateConfigEvent,
		execStatus:  configEventExecSucceed,
		description: "",
	}
	testPipelineModule := &config.PipelineModule{}
	err := updatePipeline(testPipelineModule)
	if err != nil {
		t.Errorf("test update pipeline failed %s", err.Error())
	}

	GetConfigManager().eventCallbackChan <- &configCallbackEvent{
		eventType:   updateConfigEvent,
		execStatus:  configEventExecFailed,
		description: "",
	}
	_ = updatePipeline(testPipelineModule)
}
