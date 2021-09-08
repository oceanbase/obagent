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

	"github.com/oceanbase/obagent/config"
)

//ConfigManager responsible for managing and configuring the corresponding pipeline module
type ConfigManager struct {
	configMap         map[string]*config.PipelineModule
	eventChan         chan *configEvent
	eventCallbackChan chan *configCallbackEvent
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
	eventType      configEventType
	pipelineModule *config.PipelineModule
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
			configMap:         make(map[string]*config.PipelineModule, 16),
			eventChan:         make(chan *configEvent, 16),
			eventCallbackChan: make(chan *configCallbackEvent, 16),
		}
	})
	return configMgr
}

//Schedule responsible for handling events
func (c *ConfigManager) Schedule(context context.Context) {
	go c.schedule(context)
}

func (c *ConfigManager) schedule(context context.Context) {
	for {
		select {
		case event, ok := <-c.eventChan:
			if !ok {
				log.Info("config manager event chan is closed")
				return
			}

			callbackEvent := &configCallbackEvent{
				eventType: event.eventType,
			}

			switch event.eventType {
			case addConfigEvent:

				err := c.handleAddEvent(event, callbackEvent)
				if err != nil {
					log.WithError(err).Error("config manager handle add event failed")
				}

			case deleteConfigEvent:

				err := c.handleDelEvent(event, callbackEvent)
				if err != nil {
					log.WithError(err).Error("config manager handle delete event failed")
				}

			case updateConfigEvent:

				err := c.handleUpdateEvent(event, callbackEvent)
				if err != nil {
					log.WithError(err).Error("config manager handle update event failed")
				}

			}

			c.eventCallbackChan <- callbackEvent

		case <-context.Done():
			log.Info("config manager schedule exit")
			return
		}
	}
}

func (c *ConfigManager) getConfig(module string) (*config.PipelineModule, bool) {
	pipelineModule, exist := c.configMap[module]
	return pipelineModule, exist
}

func (c *ConfigManager) setConfig(module string, pipelineModule *config.PipelineModule) {
	c.configMap[module] = pipelineModule
}

func (c *ConfigManager) delConfig(module string) {
	delete(c.configMap, module)
}

//handleAddEvent handle the add event and return.
//If error is nil, indicates that the addition was successful.
func (c *ConfigManager) handleAddEvent(event *configEvent, callbackEvent *configCallbackEvent) error {

	_, exist := c.getConfig(event.pipelineModule.Name)
	if exist {

		callbackEvent.execStatus = configEventExecFailed
		callbackEvent.description = fmt.Sprintf("moudle %s config add failed", event.pipelineModule.Name)

		return errors.Errorf("moudle %s config is exist", event.pipelineModule.Name)
	}

	status := c.handleConfigEvent(event, addPipelineEvent, callbackEvent)
	if status {
		c.setConfig(event.pipelineModule.Name, event.pipelineModule)
	}

	return nil
}

//handleDelEvent handle the delete event and return.
//If error is nil, indicates that the addition was successful.
func (c *ConfigManager) handleDelEvent(event *configEvent, callbackEvent *configCallbackEvent) error {

	_, exist := c.getConfig(event.pipelineModule.Name)
	if !exist {

		callbackEvent.execStatus = configEventExecFailed
		callbackEvent.description = fmt.Sprintf("moudle %s config delete failed", event.pipelineModule.Name)

		return errors.Errorf("moudle %s config is not exist", event.pipelineModule.Name)
	}

	status := c.handleConfigEvent(event, deletePipelineEvent, callbackEvent)
	if status {
		c.delConfig(event.pipelineModule.Name)
	}

	return nil
}

//handleUpdateEvent handle the update event and return.
//If error is nil, indicates that the addition was successful.
func (c *ConfigManager) handleUpdateEvent(event *configEvent, callbackEvent *configCallbackEvent) error {

	_, exist := c.getConfig(event.pipelineModule.Name)
	if !exist {

		callbackEvent.execStatus = configEventExecFailed
		callbackEvent.description = fmt.Sprintf("moudle %s config update failed", event.pipelineModule.Name)

		return errors.Errorf("moudle %s config is not exist", event.pipelineModule.Name)
	}

	status := c.handleConfigEvent(event, updatePipelineEvent, callbackEvent)
	if status {
		c.setConfig(event.pipelineModule.Name, event.pipelineModule)
	}

	return nil
}

//handleConfigEvent package and deliver events to the pipelineManager and get execution status
func (c *ConfigManager) handleConfigEvent(event *configEvent, pipelineEventType pipelineEventType, callbackEvent *configCallbackEvent) bool {
	pipeCallback := GetPipelineManager().handlePipelineEvent(pipelineEventType, event.pipelineModule.Name, CreatePipelineInstances(event.pipelineModule))
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
