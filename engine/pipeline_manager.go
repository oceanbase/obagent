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
	"fmt"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//PipelineManager responsible for managing the pipeline corresponding to the module name
type PipelineManager struct {
	pipelinesMap      map[string][]*PipelineInstance
	pipelineEventChan chan *pipelineEvent
	eventCallbackChan chan *pipelineCallbackEvent
}

type pipelineEventType string

type pipelineEventExecStatusType string

const (
	addPipelineEvent    pipelineEventType = "add"
	deletePipelineEvent pipelineEventType = "delete"
	updatePipelineEvent pipelineEventType = "update"

	pipelineEventExecSucceed pipelineEventExecStatusType = "succeed"
	pipelineEventExecFailed  pipelineEventExecStatusType = "failed"
)

type pipelineEvent struct {
	eventType pipelineEventType
	name      string
	pipelines []*PipelineInstance
}

type pipelineCallbackEvent struct {
	eventType   pipelineEventType
	execStatus  pipelineEventExecStatusType
	description string
}

var pipelineMgr *PipelineManager
var pipelineManagerOnce sync.Once

// GetPipelineManager get pipeline manager singleton
func GetPipelineManager() *PipelineManager {
	pipelineManagerOnce.Do(func() {
		pipelineMgr = &PipelineManager{
			pipelinesMap:      make(map[string][]*PipelineInstance, 16),
			pipelineEventChan: make(chan *pipelineEvent, 16),
			eventCallbackChan: make(chan *pipelineCallbackEvent, 16),
		}
	})
	return pipelineMgr
}

//handlePipelineEvent handle pipeline events and get callback results
func (p *PipelineManager) handlePipelineEvent(eventType pipelineEventType, name string, pipelines []*PipelineInstance) *pipelineCallbackEvent {
	event := &pipelineEvent{
		eventType: eventType,
		name:      name,
		pipelines: pipelines,
	}
	p.pipelineEventChan <- event
	return <-p.eventCallbackChan
}

//Schedule responsible for handling events
func (p *PipelineManager) Schedule(context context.Context) {
	go p.schedule(context)
}

func (p *PipelineManager) schedule(context context.Context) {
	for {
		select {
		case event, ok := <-p.pipelineEventChan:
			if !ok {
				log.Info("pipeline manager event chan closed")
				return
			}

			callbackEvent := &pipelineCallbackEvent{
				eventType: event.eventType,
			}

			switch event.eventType {
			case addPipelineEvent:

				err := p.handleAddEvent(event)
				if err != nil {
					callbackEvent.execStatus = pipelineEventExecFailed
					callbackEvent.description = fmt.Sprintf("moudle %s pipelines add failed", event.name)
					log.WithError(err).Error("pipeline manager handle add event failed")
				} else {
					callbackEvent.execStatus = pipelineEventExecSucceed
					callbackEvent.description = ""
				}

			case deletePipelineEvent:

				err := p.handleDelEvent(event)
				if err != nil {
					callbackEvent.execStatus = pipelineEventExecFailed
					callbackEvent.description = fmt.Sprintf("moudle %s pipelines delete failed", event.name)
					log.WithError(err).Error("pipeline manager handle delete event failed")
				} else {
					callbackEvent.execStatus = pipelineEventExecSucceed
					callbackEvent.description = ""
				}

			case updatePipelineEvent:

				err := p.handleUpdateEvent(event)
				if err != nil {
					callbackEvent.execStatus = pipelineEventExecFailed
					callbackEvent.description = fmt.Sprintf("moudle %s pipelines update failed", event.name)
					log.WithError(err).Error("pipeline manager handle update event failed")
				} else {
					callbackEvent.execStatus = pipelineEventExecSucceed
					callbackEvent.description = ""
				}

			}

			p.eventCallbackChan <- callbackEvent

		case <-context.Done():
			log.Info("pipeline manager scheduler exit")
			return
		}
	}
}

var initPipelinesFunc = func(pipelines []*PipelineInstance) error {
	for _, pipeline := range pipelines {
		err := pipeline.Init()
		if err != nil {
			return errors.Wrapf(err, "create pipeline %s", pipeline.name)
		}
	}
	return nil
}

var startPipelinesFunc = func(pipelines []*PipelineInstance) {
	for _, pipeline := range pipelines {
		pipeline.Start()
	}
}

var stopPipelinesFunc = func(pipelines []*PipelineInstance) {
	for _, pipeline := range pipelines {
		pipeline.Stop()
	}
}

//handleAddEvent handle the add event and return.
//If error is nil, indicates that the addition was successful.
func (p *PipelineManager) handleAddEvent(event *pipelineEvent) error {

	_, exist := p.getPipelines(event.name)
	if exist {
		return errors.Errorf("get moudle %s pipleines is exist", event.name)
	}

	err := initPipelinesFunc(event.pipelines)
	if err != nil {
		return errors.Wrapf(err, "create moudle %s pipelines", event.name)
	}

	startPipelinesFunc(event.pipelines)

	p.setPipelines(event.name, event.pipelines)

	return nil
}

//handleUpdateEvent handle the update event and return.
//If error is nil, indicates that the addition was successful.
func (p *PipelineManager) handleUpdateEvent(event *pipelineEvent) error {

	oldPipelines, exist := p.getPipelines(event.name)
	if !exist {
		return errors.Errorf("get moudle %s pipleines is not exist", event.name)
	}

	err := initPipelinesFunc(event.pipelines)
	if err != nil {
		return errors.Wrapf(err, "create moudle %s pipelines failed", event.name)
	}

	stopPipelinesFunc(oldPipelines)

	startPipelinesFunc(event.pipelines)

	p.setPipelines(event.name, event.pipelines)

	return nil
}

//handleDelEvent handle the delete event and return.
//If error is nil, indicates that the addition was successful.
func (p *PipelineManager) handleDelEvent(event *pipelineEvent) error {

	oldPipelines, exist := p.getPipelines(event.name)
	if !exist {
		return errors.Errorf("moudle %s pipleines is not exist", event.name)
	}

	stopPipelinesFunc(oldPipelines)

	p.delPipelines(event.name)

	return nil
}

func (p *PipelineManager) getPipelines(name string) ([]*PipelineInstance, bool) {
	pipelines, exist := p.pipelinesMap[name]
	return pipelines, exist
}

func (p *PipelineManager) setPipelines(name string, pipelines []*PipelineInstance) {
	p.pipelinesMap[name] = pipelines
}

func (p *PipelineManager) delPipelines(name string) {
	delete(p.pipelinesMap, name)
}
