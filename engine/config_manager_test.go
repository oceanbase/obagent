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
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/oceanbase/obagent/config"
)

func TestConfigManagerHandleEvent(t *testing.T) {
	testPipelineModule := &config.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}
	addEvent := &configEvent{
		eventType:      addConfigEvent,
		pipelineModule: testPipelineModule,
	}
	updateEvent := &configEvent{
		eventType:      updateConfigEvent,
		pipelineModule: testPipelineModule,
	}
	delEvent := &configEvent{
		eventType:      deleteConfigEvent,
		pipelineModule: testPipelineModule,
	}

	tests := []struct {
		name   string
		fields *ConfigManager
		args   *configEvent
	}{
		{name: "testAdd", fields: GetConfigManager(), args: addEvent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigManager{
				configMap:         tt.fields.configMap,
				eventChan:         tt.fields.eventChan,
				eventCallbackChan: tt.fields.eventCallbackChan,
			}

			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   addPipelineEvent,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			callback := &configCallbackEvent{}
			err = c.handleAddEvent(tt.args, callback)
			if err != nil {
				t.Errorf("config managaer test handle add event failed %s", err.Error())
			}

			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   addPipelineEvent,
				execStatus:  pipelineEventExecFailed,
				description: "",
			}
			callback = &configCallbackEvent{}
			_ = c.handleAddEvent(tt.args, callback)
		})
	}

	tests = []struct {
		name   string
		fields *ConfigManager
		args   *configEvent
	}{
		{name: "testUpdate", fields: GetConfigManager(), args: updateEvent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigManager{
				configMap:         tt.fields.configMap,
				eventChan:         tt.fields.eventChan,
				eventCallbackChan: tt.fields.eventCallbackChan,
			}
			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   updatePipelineEvent,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			callback := &configCallbackEvent{}
			err = c.handleUpdateEvent(tt.args, callback)
			if err != nil {
				t.Errorf("config managaer test handle update event failed %s", err.Error())
			}

			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   updatePipelineEvent,
				execStatus:  pipelineEventExecFailed,
				description: "",
			}
			callback = &configCallbackEvent{}
			_ = c.handleUpdateEvent(tt.args, callback)
		})
	}

	tests = []struct {
		name   string
		fields *ConfigManager
		args   *configEvent
	}{
		{name: "testDel", fields: GetConfigManager(), args: delEvent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigManager{
				configMap:         tt.fields.configMap,
				eventChan:         tt.fields.eventChan,
				eventCallbackChan: tt.fields.eventCallbackChan,
			}
			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   deletePipelineEvent,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			callback := &configCallbackEvent{}
			err = c.handleDelEvent(tt.args, callback)
			if err != nil {
				t.Errorf("config managaer test handle update event failed %s", err.Error())
			}

			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   deletePipelineEvent,
				execStatus:  pipelineEventExecFailed,
				description: "",
			}
			callback = &configCallbackEvent{}
			_ = c.handleDelEvent(tt.args, callback)
		})
	}
}

func TestConfigManagerSchedule(t *testing.T) {
	testPipelineModule := &config.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}
	tests := []struct {
		name   string
		fields *ConfigManager
	}{
		{name: "test", fields: GetConfigManager()},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigManager{
				configMap:         tt.fields.configMap,
				eventChan:         tt.fields.eventChan,
				eventCallbackChan: tt.fields.eventCallbackChan,
			}
			c.Schedule(ctx)
			event := &configEvent{
				eventType:      addConfigEvent,
				pipelineModule: testPipelineModule,
			}
			GetConfigManager().eventChan <- event
			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   addPipelineEvent,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			_ = <-GetConfigManager().eventCallbackChan

			GetConfigManager().eventChan <- event
			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   addPipelineEvent,
				execStatus:  pipelineEventExecFailed,
				description: "",
			}
			_ = <-GetConfigManager().eventCallbackChan

			event = &configEvent{
				eventType:      deleteConfigEvent,
				pipelineModule: testPipelineModule,
			}
			GetConfigManager().eventChan <- event
			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   deletePipelineEvent,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			_ = <-GetConfigManager().eventCallbackChan

			GetConfigManager().eventChan <- event
			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   deletePipelineEvent,
				execStatus:  pipelineEventExecFailed,
				description: "",
			}
			_ = <-GetConfigManager().eventCallbackChan

			event = &configEvent{
				eventType:      updateConfigEvent,
				pipelineModule: testPipelineModule,
			}
			GetConfigManager().eventChan <- event
			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   updatePipelineEvent,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			_ = <-GetConfigManager().eventCallbackChan

			GetConfigManager().eventChan <- event
			GetPipelineManager().eventCallbackChan <- &pipelineCallbackEvent{
				eventType:   updatePipelineEvent,
				execStatus:  pipelineEventExecFailed,
				description: "",
			}
			_ = <-GetConfigManager().eventCallbackChan
		})
	}
}

func TestConfigManagerConfig(t *testing.T) {
	testPipelineModule := &config.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}
	t.Run("setConfigTest", func(t *testing.T) {
		GetConfigManager().setConfig("test", testPipelineModule)
	})

	t.Run("getConfigTest", func(t *testing.T) {
		GetConfigManager().getConfig("test")
	})

	t.Run("setConfigTest", func(t *testing.T) {
		GetConfigManager().delConfig("test")
	})
}

func TestGetConfigManager(t *testing.T) {
	tests := []struct {
		name string
		want *ConfigManager
	}{
		{name: "test", want: GetConfigManager()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfigManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfigManager() = %v, want %v", got, tt.want)
			}
		})
	}
}
