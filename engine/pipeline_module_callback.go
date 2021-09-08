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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
)

func addPipeline(pipelineModule *config.PipelineModule) error {
	configManager := GetConfigManager()
	event := &configEvent{
		eventType:      addConfigEvent,
		pipelineModule: pipelineModule,
	}

	configManager.eventChan <- event
	callbackEvent := <-configManager.eventCallbackChan

	var err error
	log.Infof("add pipeline module result %s %s ", callbackEvent.execStatus, callbackEvent.description)
	switch callbackEvent.execStatus {
	case configEventExecSucceed:
		err = nil
	case configEventExecFailed:
		err = errors.Errorf("add pipeline module failed description %s", callbackEvent.description)
	}
	return err
}

func deletePipeline(pipelineModule *config.PipelineModule) error {
	configManager := GetConfigManager()
	event := &configEvent{
		eventType:      deleteConfigEvent,
		pipelineModule: pipelineModule,
	}
	log.Infof("update pipeline event %v", event)
	configManager.eventChan <- event
	callbackEvent := <-configManager.eventCallbackChan

	var err error
	log.Infof("delete pipeline module result %s %s ", callbackEvent.execStatus, callbackEvent.description)
	switch callbackEvent.execStatus {
	case configEventExecSucceed:
		err = nil
	case configEventExecFailed:
		err = errors.Errorf("delete pipeline module failed description %s", callbackEvent.description)
	}
	return err
}

func updatePipeline(pipelineModule *config.PipelineModule) error {
	configManager := GetConfigManager()
	event := &configEvent{
		eventType:      updateConfigEvent,
		pipelineModule: pipelineModule,
	}
	configManager.eventChan <- event
	callbackEvent := <-configManager.eventCallbackChan

	var err error
	log.Infof("update pipeline module result %s %s ", callbackEvent.execStatus, callbackEvent.description)
	switch callbackEvent.execStatus {
	case configEventExecSucceed:
		err = nil
	case configEventExecFailed:
		err = errors.Errorf("update pipeline module failed description %s", callbackEvent.description)
	}
	return err
}

func updateOrAddPipeline(pipelineModule *config.PipelineModule) error {
	err := updatePipeline(pipelineModule)
	if err != nil {
		log.WithError(err).Error("update or add pipeline module failed")
	} else {
		return err
	}
	err = addPipeline(pipelineModule)
	if err != nil {
		log.WithError(err).Error("update or add pipeline module failed")
	}
	return err
}

func NotifyServerBasicAuth(basicConf config.BasicAuthConfig) error {
	monagentServer := GetMonitorAgentServer()
	monagentServer.Server.BasicAuthorizer.SetConf(basicConf)
	monagentServer.Server.UseBasicAuth()
	return nil
}

func NotifyAdminServerBasicAuth(basicConf config.BasicAuthConfig) error {
	monagentServer := GetMonitorAgentServer()
	monagentServer.AdminServer.BasicAuthorizer.SetConf(basicConf)
	monagentServer.AdminServer.UseBasicAuth()
	return nil
}

func InitPipelineModuleCallback(pipelineModule *config.PipelineModule) error {
	var err error
	if pipelineModule.Status == config.INACTIVE {
		log.Warnf("pipeline module %s is inactive, just skip", pipelineModule.Name)
	} else {
		err = addPipeline(pipelineModule)
	}
	return errors.Wrap(err, "init pipeline module callback")
}

func UpdatePipelineModuleCallback(pipelineModule *config.PipelineModule) error {
	var err error
	if pipelineModule.Status == config.INACTIVE {
		err = deletePipeline(pipelineModule)
	} else {
		err = updateOrAddPipeline(pipelineModule)
	}
	return errors.Wrap(err, "update pipeline module callback")
}
