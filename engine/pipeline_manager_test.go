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
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obagent/api/web"
	"github.com/oceanbase/obagent/config"
)

func TestPipelineManagerSchedule(t *testing.T) {
	testPipelineModule := &config.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}
	tests := []struct {
		name   string
		fields *PipelineManager
	}{

		{name: "test", fields: GetPipelineManager()},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	GetPipelineManager().Schedule(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PipelineManager{
				pipelinesMap:      tt.fields.pipelinesMap,
				pipelineEventChan: tt.fields.pipelineEventChan,
				eventCallbackChan: tt.fields.eventCallbackChan,
			}

			_ = p.handlePipelineEvent(addPipelineEvent, "test", CreatePipelineInstances(testPipelineModule))

			_ = p.handlePipelineEvent(updatePipelineEvent, "test", CreatePipelineInstances(testPipelineModule))

			_ = p.handlePipelineEvent(deletePipelineEvent, "test", CreatePipelineInstances(testPipelineModule))

		})
	}
}

func TestPipelineManagerHandleEvent(t *testing.T) {
	testPipelineModule := &config.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}
	fmt.Println(testPipelineModule.Pipelines[1].Config)
	pipelineInstances := CreatePipelineInstances(testPipelineModule)
	addEvent := &pipelineEvent{
		eventType: addPipelineEvent,
		name:      "test",
		pipelines: pipelineInstances,
	}
	updateEvent := &pipelineEvent{
		eventType: updatePipelineEvent,
		name:      "test",
		pipelines: pipelineInstances,
	}
	delEvent := &pipelineEvent{
		eventType: deletePipelineEvent,
		name:      "test",
		pipelines: pipelineInstances,
	}

	tests := []struct {
		name   string
		fields *PipelineManager
		args   *pipelineEvent
	}{
		{name: "testAdd", fields: GetPipelineManager(), args: addEvent},
	}

	monitorAgentServer = &MonitorAgentServer{}
	GetMonitorAgentServer().Server = &web.HttpServer{
		Router: gin.Default(),
		Server: &http.Server{
			Addr: ":0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PipelineManager{
				pipelinesMap:      tt.fields.pipelinesMap,
				pipelineEventChan: tt.fields.pipelineEventChan,
				eventCallbackChan: tt.fields.eventCallbackChan,
			}

			_ = p.handleAddEvent(tt.args)
		})
	}

	tests = []struct {
		name   string
		fields *PipelineManager
		args   *pipelineEvent
	}{
		{name: "testUpdate", fields: GetPipelineManager(), args: updateEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PipelineManager{
				pipelinesMap:      tt.fields.pipelinesMap,
				pipelineEventChan: tt.fields.pipelineEventChan,
				eventCallbackChan: tt.fields.eventCallbackChan,
			}

			_ = p.handleUpdateEvent(tt.args)
		})
	}

	tests = []struct {
		name   string
		fields *PipelineManager
		args   *pipelineEvent
	}{
		{name: "testDelete", fields: GetPipelineManager(), args: delEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PipelineManager{
				pipelinesMap:      tt.fields.pipelinesMap,
				pipelineEventChan: tt.fields.pipelineEventChan,
				eventCallbackChan: tt.fields.eventCallbackChan,
			}

			_ = p.handleDelEvent(tt.args)
		})
	}
}

func TestPipelineManagerPipelines(t *testing.T) {
	testPipelineModule := &config.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}
	pipelineInstances := CreatePipelineInstances(testPipelineModule)
	t.Run("setConfigTest", func(t *testing.T) {
		GetPipelineManager().setPipelines("test", pipelineInstances)
	})

	t.Run("getConfigTest", func(t *testing.T) {
		GetPipelineManager().getPipelines("test")
	})

	t.Run("setConfigTest", func(t *testing.T) {
		GetPipelineManager().delPipelines("test")
	})
}
