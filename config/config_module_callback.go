package config

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ModuleType string

var (
	// module config callbacks
	callbacks map[ModuleType]*ConfigCallback
	modules   map[string]ModuleType
)

// init callbacks and register module config callbacks
func init() {
	callbacks = make(map[ModuleType]*ConfigCallback, 8)
	modules = make(map[string]ModuleType, 32)
}

// Callback call Callback when receive a request
type Callback func(ctx context.Context, nconfig *NotifyModuleConfig) error

// ModuleCallback register a ModuleCallback, Callback will call it
type ModuleCallback func(ctx context.Context, moduleConf interface{}) error

// Creator create a new module config
type Creator func() interface{}

// ConfigCallback module config consumer, call updateConfigCallback to update config,
// call InitConfigCallback to check if config is identical
type ConfigCallback struct {
	moduleType           ModuleType
	creator              Creator
	InitConfigCallback   func(ctx context.Context, module string) error
	NotifyConfigCallback Callback
}

// create instance for module type
func createModuleTypeInstance(module string) (interface{}, error) {
	moduleType, ex := modules[module]
	if !ex {
		return nil, errors.Errorf("module %s's moduleType is not found.", module)
	}
	callback, ex := getModuleTypeCallback(moduleType)
	if !ex {
		return nil, errors.Errorf("module %s moduleType %s callback is not exist!", module, moduleType)
	}

	return callback.creator(), nil
}

func getModuleCallback(module string) (callback *ConfigCallback, exists bool) {
	moduleType, ex := modules[module]
	if !ex {
		log.Errorf("module %s 's module type is not found", module)
		return nil, false
	}
	return getModuleTypeCallback(moduleType)
}

func getModuleTypeCallback(moduleType ModuleType) (callback *ConfigCallback, exists bool) {
	callback, exists = callbacks[moduleType]
	return
}

func RegisterModule(module string, moduleType ModuleType) error {
	existModuleType, ex := modules[module]
	if ex && existModuleType != moduleType {
		return errors.Errorf("module %s 's module type %s is already set as %s", module, moduleType, existModuleType)
	}
	modules[module] = moduleType
	log.Debugf("register module %s, moduleType %s", module, moduleType)
	return nil
}

// RegisterConfigCallback registry module config consumer
func RegisterConfigCallback(moduleType ModuleType,
	creator Creator,
	initConfig ModuleCallback,
	receiveUpdatedConfig ModuleCallback) error {

	log.Debugf("register moduleType %s config callback start", moduleType)

	if _, ex := getModuleTypeCallback(moduleType); ex {
		return errors.Errorf("moduleType %s callback has been registred!", moduleType)
	}

	initConfigCallback := func(ctx context.Context, module string) error {
		moduleConfigSample, ex := GetModuleConfigs()[module]
		if !ex {
			return errors.Errorf("module %s config sample is not found", module)
		}
		log.WithContext(ctx).Debugf("module %s config sample:%+v", module, moduleConfigSample)
		if moduleConfigSample.Disabled {
			return errors.Errorf("module %s config is disabled.", module)
		}

		moduleConf, err := getFinalModuleConfig(module, moduleConfigSample.Config, nil)
		if err != nil {
			return errors.Errorf("get module %s final config err:%s", module, err)
		}

		err = initConfig(ctx, moduleConf)
		if err != nil {
			return errors.Errorf("init module %s conf err:%s", module, err)
		}
		return nil
	}

	NotifyUpdatedConfigCallback := func(ctx context.Context, econfig *NotifyModuleConfig) error {
		log.WithContext(ctx).Infof("notify module %s config, process:%s", econfig.Module, econfig.Process)
		err := receiveUpdatedConfig(ctx, econfig.Config)
		if err != nil {
			return errors.Errorf("notify module %s config, process %s, err:%s", econfig.Module, econfig.Process, err)
		}
		return nil
	}

	callbacks[moduleType] = &ConfigCallback{
		moduleType:           moduleType,
		creator:              creator,
		InitConfigCallback:   initConfigCallback,
		NotifyConfigCallback: NotifyUpdatedConfigCallback,
	}

	log.Debugf("register moduleType %s config callback end", moduleType)
	return nil
}
