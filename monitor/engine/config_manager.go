package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	agentlog "github.com/oceanbase/obagent/log"
)

// ConfigManager responsible for managing and configuring the corresponding pipeline module
type ConfigManager struct {
	configMap map[string]*monagent.PipelineModule
	eventChan chan *configEvent
}

type configEventType string

type configEventExecStatusType string

const (
	addConfigEvent configEventType = "add"

	deleteConfigEvent configEventType = "delete"

	updateConfigEvent configEventType = "update"

	configEventExecSucceed configEventExecStatusType = "succeed"

	configEventExecFailed configEventExecStatusType = "failed"
)

type configEvent struct {
	ctx            context.Context
	eventType      configEventType
	pipelineModule *monagent.PipelineModule
	callbackChan   chan *configCallbackEvent
}

type configCallbackEvent struct {
	eventType   configEventType
	execStatus  configEventExecStatusType
	description string
}

var configMgr *ConfigManager
var configManagerOnce sync.Once

// GetConfigManager get config manager singleton
func GetConfigManager() *ConfigManager {
	configManagerOnce.Do(func() {
		configMgr = &ConfigManager{
			configMap: make(map[string]*monagent.PipelineModule, 16),
			eventChan: make(chan *configEvent, 16),
		}
	})
	return configMgr
}

// Schedule responsible for handling events
func (c *ConfigManager) Schedule(context context.Context) {
	go c.schedule(context)
}

func (c *ConfigManager) schedule(ctx context.Context) {
	for {
		select {
		case event, ok := <-c.eventChan:
			if !ok {
				log.WithContext(event.ctx).Info("config manager event chan is closed")
				return
			}
			logger := log.WithContext(context.WithValue(event.ctx, agentlog.StartTimeKey, time.Now())).WithField("module", event.pipelineModule.Name).WithField("eventType", event.eventType)
			logger.Info("receive module config")

			callbackEvent := &configCallbackEvent{
				eventType: event.eventType,
			}

			switch event.eventType {
			case addConfigEvent:

				err := c.handleAddEvent(event, callbackEvent)
				if err != nil {
					logger.WithError(err).Error("config manager handle add event failed")
				}

			case deleteConfigEvent:

				err := c.handleDelEvent(event, callbackEvent)
				if err != nil {
					logger.WithError(err).Error("config manager handle delete event failed")
				}

			case updateConfigEvent:

				err := c.handleUpdateEvent(event, callbackEvent)
				if err != nil {
					logger.WithError(err).Error("config manager handle update event failed")
				}

			}

			logger.Infof("module config event handle compeleted")
			event.callbackChan <- callbackEvent

		case <-ctx.Done():
			log.Info("config manager schedule exit")
			return
		}
	}
}

func (c *ConfigManager) getConfig(module string) (*monagent.PipelineModule, bool) {
	pipelineModule, exist := c.configMap[module]
	return pipelineModule, exist
}

func (c *ConfigManager) setConfig(module string, pipelineModule *monagent.PipelineModule) {
	c.configMap[module] = pipelineModule
}

func (c *ConfigManager) delConfig(module string) {
	delete(c.configMap, module)
}

// handleAddEvent handle the add event and return.
// If error is nil, indicates that the addition was successful.
func (c *ConfigManager) handleAddEvent(event *configEvent, callbackEvent *configCallbackEvent) error {

	_, exist := c.getConfig(event.pipelineModule.Name)
	if exist {
		callbackEvent.execStatus = configEventExecFailed
		callbackEvent.description = fmt.Sprintf("module %s config already exists", event.pipelineModule.Name)
		return errors.Errorf("module %s config already exists", event.pipelineModule.Name)
	}

	status := c.handleConfigEvent(event, addPipelineEvent, callbackEvent)
	if status {
		c.setConfig(event.pipelineModule.Name, event.pipelineModule)
	}

	return nil
}

// handleDelEvent handle the delete event and return.
// If error is nil, indicates that the addition was successful.
func (c *ConfigManager) handleDelEvent(event *configEvent, callbackEvent *configCallbackEvent) error {

	_, exist := c.getConfig(event.pipelineModule.Name)
	if !exist {
		callbackEvent.execStatus = configEventExecSucceed
		callbackEvent.description = fmt.Sprintf("module %s config not exists", event.pipelineModule.Name)
		return nil
	}

	status := c.handleConfigEvent(event, deletePipelineEvent, callbackEvent)
	if status {
		c.delConfig(event.pipelineModule.Name)
	}

	return nil
}

// handleUpdateEvent handle the update event and return.
// If error is nil, indicates that the addition was successful.
func (c *ConfigManager) handleUpdateEvent(event *configEvent, callbackEvent *configCallbackEvent) error {

	_, exist := c.getConfig(event.pipelineModule.Name)
	if !exist {
		log.WithContext(event.ctx).Warnf("update module %s config not exist.", event.pipelineModule.Name)
	}

	status := c.handleConfigEvent(event, updatePipelineEvent, callbackEvent)
	if status {
		c.setConfig(event.pipelineModule.Name, event.pipelineModule)
	}

	return nil
}

// handleConfigEvent package and deliver events to the pipelineManager and get execution status
func (c *ConfigManager) handleConfigEvent(event *configEvent, pipelineEventType pipelineEventType, callbackEvent *configCallbackEvent) bool {
	logger := log.WithContext(context.WithValue(event.ctx, agentlog.StartTimeKey, time.Now())).WithField("module", event.pipelineModule.Name)
	pipelines, err := CreatePipelines(event.pipelineModule)
	if err != nil {
		log.WithContext(event.ctx).WithError(err).Error("CreatePipelines failed")
		callbackEvent.execStatus = configEventExecFailed
		callbackEvent.description = "CreatePipelines failed"
		return false
	}
	logger.Infof("pipeline event is created")

	pipelineEvent := &pipelineEvent{
		ctx:       event.ctx,
		eventType: pipelineEventType,
		name:      event.pipelineModule.Name,
		//pipelines:    CreatePipelineInstances(event.pipelineModule),
		pipelines:    pipelines,
		callbackChan: make(chan *pipelineCallbackEvent, 1),
	}

	pipeCallback := GetPipelineManager().handlePipelineEvent(pipelineEvent)
	logger.Infof("handle pipeline event compeleted")

	var execStatus bool
	switch pipeCallback.execStatus {
	case pipelineEventExecSucceed:
		callbackEvent.execStatus = configEventExecSucceed
		callbackEvent.description = pipeCallback.description
		execStatus = true
	case pipelineEventExecFailed:
		callbackEvent.execStatus = configEventExecFailed
		callbackEvent.description = pipeCallback.description
		execStatus = false
	}
	return execStatus
}
