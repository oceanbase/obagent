package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	errors2 "github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/goroutinepool"
	agentlog "github.com/oceanbase/obagent/log"
)

// PipelineManager responsible for managing the pipeline corresponding to the module name
type PipelineManager struct {
	lock sync.Mutex
	//pipelinesMap      map[string][]*PipelineInstance
	pipelinesMap      map[string][]*Pipeline
	pipelineEventChan chan *pipelineEvent
	eventTaskPool     *goroutinepool.GoroutinePool
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
	ctx       context.Context
	eventType pipelineEventType
	name      string
	//pipelines    []*PipelineInstance
	pipelines    []*Pipeline
	callbackChan chan *pipelineCallbackEvent
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
		eventTaskPool, err := goroutinepool.NewGoroutinePool("PipelineManager", 4, 32)
		if err != nil {
			log.Fatal(err)
		}
		pipelineMgr = &PipelineManager{
			pipelinesMap:      make(map[string][]*Pipeline, 16),
			pipelineEventChan: make(chan *pipelineEvent, 16),
			eventTaskPool:     eventTaskPool,
		}
	})
	return pipelineMgr
}

// handlePipelineEvent handle pipeline events and get callback results
func (p *PipelineManager) handlePipelineEvent(event *pipelineEvent) *pipelineCallbackEvent {
	p.pipelineEventChan <- event
	return <-event.callbackChan
}

// Schedule responsible for handling events
func (p *PipelineManager) Schedule(context context.Context) {
	go p.schedule(context)
}

// Stop responsible for stop all pipelines(NOTE: 因为后面 pipeline 都会迁移到 ，故这里先关闭 )
func (p *PipelineManager) Stop(context context.Context) {
	for _, pipelines := range p.pipelinesMap {
		for _, p := range pipelines {
			p.Stop()
		}
	}
}

// for test only, send call back event without really handle pipeline event
func (p *PipelineManager) emptySchedule(ctx context.Context) {
	for {
		select {
		case event, ok := <-p.pipelineEventChan:
			if !ok {
				log.Info("pipeline manager event chan closed")
				return
			}

			successCallbackEvent := &pipelineCallbackEvent{
				eventType:   event.eventType,
				execStatus:  pipelineEventExecSucceed,
				description: "",
			}
			event.callbackChan <- successCallbackEvent

		case <-ctx.Done():
			// p.eventTaskPool.Close()
			log.Info("pipeline manager scheduler exit")
			return
		}
	}
}

func (p *PipelineManager) schedule(ctx context.Context) {
	for {
		select {
		case event, ok := <-p.pipelineEventChan:
			if !ok {
				log.WithContext(event.ctx).Info("pipeline manager event chan closed")
				return
			}

			p.eventTaskPool.PutWithTimeout(fmt.Sprintf("%s-%s", event.name, event.eventType), func() error {
				p.handleEvent(event)
				return nil
			}, time.Minute)

		case <-ctx.Done():
			p.eventTaskPool.Close()
			log.Info("pipeline manager scheduler exit")
			return
		}
	}
}

func (p *PipelineManager) handleEvent(event *pipelineEvent) {
	logger := log.WithContext(context.WithValue(event.ctx, agentlog.StartTimeKey, time.Now())).WithField("module", event.name).WithField("eventType", event.eventType)
	logger.Infof("receive pipelie event")

	callbackEvent := &pipelineCallbackEvent{
		eventType: event.eventType,
	}

	switch event.eventType {
	case addPipelineEvent:

		err := p.handleAddEvent(event)
		if err != nil {
			callbackEvent.execStatus = pipelineEventExecFailed
			callbackEvent.description = fmt.Sprintf("add module %s failed, reason: %s", event.name, err)
			logger.WithError(err).Error("pipeline manager handle add event failed")
		} else {
			callbackEvent.execStatus = pipelineEventExecSucceed
			callbackEvent.description = ""
		}

	case deletePipelineEvent:

		err := p.handleDelEvent(event)
		if err != nil {
			callbackEvent.execStatus = pipelineEventExecFailed
			callbackEvent.description = fmt.Sprintf("delete module %s failed, reason: %s", event.name, err)
			logger.WithError(err).Error("pipeline manager handle delete event failed")
		} else {
			callbackEvent.execStatus = pipelineEventExecSucceed
			callbackEvent.description = ""
		}

	case updatePipelineEvent:

		err := p.handleUpdateEvent(event)
		if err != nil {
			callbackEvent.execStatus = pipelineEventExecFailed
			callbackEvent.description = fmt.Sprintf("update module %s failed, reason: %s", event.name, err)
			logger.WithError(err).Error("pipeline manager handle update event failed")
		} else {
			callbackEvent.execStatus = pipelineEventExecSucceed
			callbackEvent.description = ""
		}

	}

	logger.Infof("pipeline event handle completed")
	event.callbackChan <- callbackEvent
}

var startPipelinesFunc = func(ctx context.Context, pipelines []*Pipeline) error {
	for _, pipeline := range pipelines {
		log.WithContext(ctx).Infof("start pipeline %s", pipeline.Name)
		err := pipeline.Start(ctx)
		if err != nil {
			log.WithContext(ctx).Errorf("start pipeline %s, err:%+v", pipeline.Name, err)
			return err
		}
	}
	return nil
}

var closePipelinesFunc = func(ctx context.Context, pipelines []*Pipeline) {
	for _, pipeline := range pipelines {
		pipeline.Stop()
	}
}

// handleAddEvent handle the add event and return.
// If error is nil, indicates that the addition was successful.
func (p *PipelineManager) handleAddEvent(event *pipelineEvent) error {
	ctxLog := log.WithContext(event.ctx).WithField("pipelines", event.name)
	var err error
	{
		if len(event.pipelines) != 0 {
			_, exist := p.getPipelines(event.name)
			if !exist {
				err1 := startPipelinesFunc(event.ctx, event.pipelines)
				if err1 != nil {
					ctxLog.WithError(err1).Error("startPipelinesFunc failed")
					return err1
				}

				p.setPipelines(event.name, event.pipelines)
			} else {
				err = errors.Errorf("module %s already exist", event.name)
				ctxLog.WithError(err).Error("pipeline check exist failed")
			}
		}
	}

	return err
}

// handleUpdateEvent handle the update event and return.
// If error is nil, indicates that the addition was successful.
func (p *PipelineManager) handleUpdateEvent(event *pipelineEvent) error {
	ctxLog := log.WithContext(event.ctx).WithField("pipelines", event.name)
	var err error
	{
		if len(event.pipelines) != 0 {
			oldPipelines, exist := p.getPipelines(event.name)
			if exist {
				closePipelinesFunc(event.ctx, oldPipelines)

				err1 := startPipelinesFunc(event.ctx, event.pipelines)
				if err1 != nil {
					ctxLog.WithError(err1).Error("startPipelinesFunc failed")
					return err1
				}

				p.setPipelines(event.name, event.pipelines)
			} else {
				err = errors.Errorf("handleUpdateEvent module %s does not exist", event.name)
				ctxLog.WithError(err).Error("pipeline check exist failed")
			}
		}
	}

	return err
}

// handleDelEvent handle the delete event and return.
// If error is nil, indicates that the addition was successful.
func (p *PipelineManager) handleDelEvent(event *pipelineEvent) error {
	ctxLog := log.WithContext(event.ctx).WithField("pipelines", event.name)
	var err error
	{
		if len(event.pipelines) != 0 {
			oldPipelines, exist := p.getPipelines(event.name)
			if exist {
				closePipelinesFunc(event.ctx, oldPipelines)

				p.delPipelines(event.name)
			} else {
				err = errors.Errorf("handleDelEvent module %s does not exist", event.name)
				ctxLog.WithError(err).Error("pipeline check exist failed")
			}
		}
	}

	return err
}

func (p *PipelineManager) getPipelines(name string) ([]*Pipeline, bool) {
	p.lock.Lock()
	pipelines, exist := p.pipelinesMap[name]
	p.lock.Unlock()
	return pipelines, exist
}

func (p *PipelineManager) setPipelines(name string, pipelines []*Pipeline) {
	p.lock.Lock()
	p.pipelinesMap[name] = pipelines
	p.lock.Unlock()
}

type PipelineOperationResultInfo string

const (
	StartPipeline PipelineOperationResultInfo = "start"
	StopPipeline  PipelineOperationResultInfo = "stop"
)

func (p *PipelineManager) delPipelines(name string) {
	delete(p.pipelinesMap, name)
}

var PipelineStatus = StartPipeline

func PipelineLoad(ctx context.Context) (bool, error) {
	if PipelineStatus == StartPipeline {
		log.WithContext(ctx).Warn("pipeline is already start")
		return false, errors2.Occur(errors2.ErrMonPipelineStart)
	}
	p := GetPipelineManager()
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, pipelines := range p.pipelinesMap {
		for _, pipeline := range pipelines {
			err := pipeline.Start(ctx)
			if err != nil {
				log.WithError(err).Error("start pipeline failed")
				return false, errors2.Occur(errors2.ErrMonPipelineStartFail)
			}
		}
	}
	PipelineStatus = StartPipeline
	log.WithContext(ctx).Info("start pipeline success")
	return true, nil
}

func PipelineUnload(ctx context.Context) (bool, error) {
	if PipelineStatus == StopPipeline {
		// repeat stop operation is allowed
		log.WithContext(ctx).Warn("pipeline is already stop")
		return true, nil
	}
	p := GetPipelineManager()
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, pipelines := range p.pipelinesMap {
		for _, pipeline := range pipelines {
			pipeline.Stop()
		}
	}
	PipelineStatus = StopPipeline
	log.WithContext(ctx).Info("stop pipeline success")
	return true, nil
}
