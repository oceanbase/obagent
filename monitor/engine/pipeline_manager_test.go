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
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/oceanbase/obagent/config/monagent"
	errors2 "github.com/oceanbase/obagent/errors"
)

func TestPipelineManagerSchedule(t *testing.T) {
	testPipelineModule := &monagent.PipelineModule{}
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

	go GetPipelineManager().emptySchedule(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PipelineManager{
				pipelinesMap:      tt.fields.pipelinesMap,
				pipelineEventChan: tt.fields.pipelineEventChan,
			}
			ctx := context.Background()

			pipelines, err := CreatePipelines(testPipelineModule)
			if err != nil {
				t.Errorf("create pipelines failed %s", err.Error())
				return
			}
			event := &pipelineEvent{
				ctx:          ctx,
				eventType:    addPipelineEvent,
				name:         "test",
				pipelines:    pipelines,
				callbackChan: make(chan *pipelineCallbackEvent, 1),
			}
			successCallbackEvent := &pipelineCallbackEvent{
				eventType:   event.eventType,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			event.callbackChan <- successCallbackEvent
			_ = p.handlePipelineEvent(event)

			event = &pipelineEvent{
				ctx:          ctx,
				eventType:    updatePipelineEvent,
				name:         "test",
				pipelines:    pipelines,
				callbackChan: make(chan *pipelineCallbackEvent, 1),
			}
			successCallbackEvent = &pipelineCallbackEvent{
				eventType:   event.eventType,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			event.callbackChan <- successCallbackEvent
			_ = p.handlePipelineEvent(event)

			event = &pipelineEvent{
				ctx:          ctx,
				eventType:    deletePipelineEvent,
				name:         "test",
				pipelines:    pipelines,
				callbackChan: make(chan *pipelineCallbackEvent, 1),
			}
			successCallbackEvent = &pipelineCallbackEvent{
				eventType:   event.eventType,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			event.callbackChan <- successCallbackEvent
			_ = p.handlePipelineEvent(event)

		})
	}
}

func TestPipelineManagerHandleEvent(t *testing.T) {
	testPipelineModule := &monagent.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}
	fmt.Println(testPipelineModule.Pipelines[1].Config)
	pipelines, err := CreatePipelines(testPipelineModule)
	if err != nil {
		t.Errorf("CreatePipelineV2s failed %s", err.Error())
		return
	}
	ctx := context.Background()
	addEvent := &pipelineEvent{
		ctx:       ctx,
		eventType: addPipelineEvent,
		name:      "test",
		pipelines: pipelines,
	}
	updateEvent := &pipelineEvent{
		ctx:       ctx,
		eventType: updatePipelineEvent,
		name:      "test",
		pipelines: pipelines,
	}
	delEvent := &pipelineEvent{
		ctx:       ctx,
		eventType: deletePipelineEvent,
		name:      "test",
		pipelines: pipelines,
	}

	tests := []struct {
		name   string
		fields *PipelineManager
		args   *pipelineEvent
	}{
		{name: "testAdd", fields: GetPipelineManager(), args: addEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PipelineManager{
				pipelinesMap:      tt.fields.pipelinesMap,
				pipelineEventChan: tt.fields.pipelineEventChan,
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
			}

			_ = p.handleDelEvent(tt.args)
		})
	}
}

func TestPipelineManagerPipelines(t *testing.T) {
	testPipelineModule := &monagent.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test config manager set config json decode failed %s", err.Error())
		return
	}
	pipelineInstances, _ := CreatePipelines(testPipelineModule)
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

func TestPipelineLoad(t *testing.T) {
	ctx := context.Background()
	res1, err := PipelineLoad(ctx)
	require.Equal(t, false, res1)
	e := errors2.Occur(errors2.ErrMonPipelineStart)
	require.Equal(t, e, err)
}

func TestPipelineUnload(t *testing.T) {
	ctx := context.Background()
	res1, err := PipelineUnload(ctx)
	require.Equal(t, nil, err)
	require.True(t, res1)
}

func TestPipelineStopAndStart(t *testing.T) {
	ctx := context.Background()
	res1, err := PipelineUnload(ctx)
	require.Equal(t, nil, err)
	require.True(t, res1)

	res2, err := PipelineLoad(ctx)
	require.Equal(t, nil, err)
	require.True(t, res2)
}

func TestPipelineRepeatStop(t *testing.T) {
	ctx := context.Background()
	res1, err := PipelineUnload(ctx)
	require.Equal(t, nil, err)
	require.True(t, res1)

	res2, err := PipelineUnload(ctx)
	require.Equal(t, nil, err)
	require.True(t, res2)
}
