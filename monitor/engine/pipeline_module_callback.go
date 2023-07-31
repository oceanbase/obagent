package engine

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/lib/retry"
	agentlog "github.com/oceanbase/obagent/log"
)

var pipelineConfigRetry = map[string]*retry.Retry{}
var pipelineConfigMutex = sync.Mutex{}

func addPipeline(ctx context.Context, pipelineModule *monagent.PipelineModule) error {
	logger := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, time.Now())).WithField("module", pipelineModule.Name)
	configManager := GetConfigManager()
	event := &configEvent{
		ctx:            ctx,
		eventType:      addConfigEvent,
		pipelineModule: pipelineModule,
		callbackChan:   make(chan *configCallbackEvent, 1),
	}

	logger.Infof("add pipeline module config event")
	configManager.eventChan <- event
	callbackEvent := <-event.callbackChan

	var err error

	logger.Infof("add pipeline module config result status: %s, description: %s", callbackEvent.execStatus, callbackEvent.description)
	switch callbackEvent.execStatus {
	case configEventExecSucceed:
		err = nil
	case configEventExecFailed:
		err = errors.Errorf("add pipeline module config %s failed, description %s", pipelineModule.Name, callbackEvent.description)
	}
	return err
}

func deletePipeline(ctx context.Context, pipelineModule *monagent.PipelineModule) error {
	logger := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, time.Now())).WithField("module", pipelineModule.Name)
	configManager := GetConfigManager()
	event := &configEvent{
		ctx:            ctx,
		eventType:      deleteConfigEvent,
		pipelineModule: pipelineModule,
		callbackChan:   make(chan *configCallbackEvent, 1),
	}

	logger.Infof("delete pipeline module config event")
	configManager.eventChan <- event
	callbackEvent := <-event.callbackChan

	var err error
	logger.Infof("delete pipeline module result status: %s, description: %s", callbackEvent.execStatus, callbackEvent.description)
	switch callbackEvent.execStatus {
	case configEventExecSucceed:
		err = nil
	case configEventExecFailed:
		err = errors.Errorf("delete pipeline module failed description %s", callbackEvent.description)
	}
	return err
}

func updatePipeline(ctx context.Context, pipelineModule *monagent.PipelineModule) error {
	logger := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, time.Now())).WithField("module", pipelineModule.Name)
	configManager := GetConfigManager()
	event := &configEvent{
		ctx:            ctx,
		eventType:      updateConfigEvent,
		pipelineModule: pipelineModule,
		callbackChan:   make(chan *configCallbackEvent, 1),
	}

	logger.Infof("update pipeline module config event")
	configManager.eventChan <- event
	callbackEvent := <-event.callbackChan

	var err error
	logger.Infof("update pipeline module config result status: %s, description: %s", callbackEvent.execStatus, callbackEvent.description)
	switch callbackEvent.execStatus {
	case configEventExecSucceed:
		err = nil
	case configEventExecFailed:
		err = errors.Errorf("update pipeline module config failed description %s", callbackEvent.description)
	}
	return err
}

func updateOrAddPipeline(ctx context.Context, pipelineModule *monagent.PipelineModule) error {
	err := updatePipeline(ctx, pipelineModule)
	if err == nil {
		return nil
	}
	log.WithContext(ctx).WithError(err).Errorf("update pipeline module config failed, module name: %s", pipelineModule.Name)

	err = addPipeline(ctx, pipelineModule)
	if err != nil {
		log.WithContext(ctx).WithError(err).Errorf("add pipeline module config failed, module name: %s", pipelineModule.Name)
	}
	return err
}

func InitPipelineModuleCallback(ctx context.Context, pipelineModule *monagent.PipelineModule) error {
	if !pipelineModule.Status.Validate() {
		log.WithContext(ctx).Warnf("pipeline module config %s is invalid, just skip", pipelineModule.Name)
		return nil
	}
	if pipelineModule.Status == monagent.INACTIVE {
		log.WithContext(ctx).Warnf("pipeline module config %s is inactive or invalid, just skip", pipelineModule.Name)
	} else {
		notifyPipelineConfigRetryIfFailed(ctx, pipelineModule.Name, func(ctx context.Context) error {
			err := addPipeline(ctx, pipelineModule)
			if err != nil {
				return err
			}
			log.WithContext(ctx).Infof("add pipeline module config %s successfully.", pipelineModule.Name)
			return nil
		})
	}
	return nil
}

func UpdatePipelineModuleCallback(ctx context.Context, pipelineModule *monagent.PipelineModule) error {
	if !pipelineModule.Status.Validate() {
		log.WithContext(ctx).Warnf("pipeline module config %s is invalid, just skip", pipelineModule.Name)
		return nil
	}
	if pipelineModule.Status == monagent.INACTIVE {
		notifyPipelineConfigRetryIfFailed(ctx, pipelineModule.Name, func(ctx context.Context) error {
			err := deletePipeline(ctx, pipelineModule)
			if err != nil {
				return err
			}
			log.WithContext(ctx).Infof("delete pipeline module config %s successfully.", pipelineModule.Name)
			return nil
		})
	} else {
		notifyPipelineConfigRetryIfFailed(ctx, pipelineModule.Name, func(ctx context.Context) error {
			err := updateOrAddPipeline(ctx, pipelineModule)
			if err != nil {
				return err
			}
			log.WithContext(ctx).Infof("update or add pipeline module config %s successfully.", pipelineModule.Name)
			return nil
		})
	}
	return nil
}

func notifyPipelineConfigRetryIfFailed(ctx context.Context, pipeline string, retryEvent func(ctx context.Context) error) {
	pipelineConfigMutex.Lock()
	defer pipelineConfigMutex.Unlock()
	if _, ex := pipelineConfigRetry[pipeline]; !ex {
		pipelineConfigRetry[pipeline] = retry.NewRetry()
	}
	pipelineConfigRetry[pipeline].Do(ctx,
		retry.RetryConfig{
			Name:          pipeline,
			RetryEvent:    retryEvent,
			RetryTimeout:  time.Second * 30,
			MaxRetryTimes: 0,
			RetryDuration: &retry.DefaultNextDuration{
				BaseDuration: time.Second * 30,
				MaxDuration:  time.Minute * 10,
			}})
}
