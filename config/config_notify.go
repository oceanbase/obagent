package config

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/executor/agent"
)

var (
	processConfigNotifyAddresses     map[string]ProcessConfigNotifyAddress
	processConfigNotifyAddressesLock sync.Mutex
)

func init() {
	processConfigNotifyAddresses = make(map[string]ProcessConfigNotifyAddress, 4)
}

// ProcessConfigNotifyAddress business process configuration information.
// This information serves as sdk source data for the mgragent and agentctl.
type ProcessConfigNotifyAddress struct {
	Local         bool            `yaml:"local"`
	Process       string          `yaml:"process"`
	NotifyAddress string          `yaml:"notifyAddress"`
	AuthConfig    BasicAuthConfig `yaml:"authConfig"`
}

func getProcessModuleConfigNotifyAddress(process string) (ProcessConfigNotifyAddress, bool) {
	processConfigNotifyAddressesLock.Lock()
	defer processConfigNotifyAddressesLock.Unlock()
	notify, ex := processConfigNotifyAddresses[process]
	return notify, ex
}

func SetProcessModuleConfigNotifyAddress(notifyConfig ProcessConfigNotifyAddress) {
	processConfigNotifyAddressesLock.Lock()
	defer processConfigNotifyAddressesLock.Unlock()
	processConfigNotifyAddresses[notifyConfig.Process] = notifyConfig
}

// InitModuleConfig init module config, then the init callback will be trigger
func InitModuleConfig(ctx context.Context, module string) error {
	callback, ex := getModuleCallback(module)
	if !ex {
		return errors.Errorf("module %s callback is not found", module)
	}
	return callback.InitConfigCallback(ctx, module)
}

// InitModuleTypeConfig init module type config, then the init callback will be trigger
func InitModuleTypeConfig(ctx context.Context, moduleType ModuleType) error {
	var errs []string
	for m, t := range modules {
		if t == moduleType {
			err := InitModuleConfig(ctx, m)
			if err != nil {
				errs = append(errs, errors.Errorf("init module %s err:%s", m, err).Error())
			}
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ","))
	}
	return nil
}

// NotifyModuleConfigs notify module config changed
func NotifyModuleConfigs(ctx context.Context, verifyConfigResult *VerifyConfigResult) error {
	if verifyConfigResult == nil {
		return nil
	}
	log.WithContext(ctx).Infof("notify module configs length:%d", len(verifyConfigResult.UpdatedConfigs))
	var errs []string
	for _, conf := range verifyConfigResult.UpdatedConfigs {
		err := notifyModuleConfig(ctx, conf)
		if err != nil {
			log.WithContext(ctx).Errorf("notify module %s err:%+v", conf.Module, err)
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.Errorf("notify modules err:%s", strings.Join(errs, ","))
	}
	return nil
}

// notify module config changed
func notifyModuleConfig(ctx context.Context, econfig *NotifyModuleConfig) error {
	log.WithContext(ctx).Infof("module %s, notify process %s, current process %s", econfig.Module, econfig.Process, CurProcess)
	notifyAddress, ex := getProcessModuleConfigNotifyAddress(string(econfig.Process))
	if !ex {
		return errors.Errorf("process %s notify config is not found", econfig.Process)
	}
	if !notifyAddress.Local && econfig.Process != CurProcess {
		return NotifyRemoteModuleConfig(ctx, econfig)
	}
	return NotifyLocalModuleConfig(ctx, econfig)
}

// NotifyLocalModuleConfig notify local process's module config changed
func NotifyLocalModuleConfig(ctx context.Context, econfig *NotifyModuleConfig) error {
	callback, ex := getModuleCallback(econfig.Module)
	if !ex {
		return errors.Errorf("module %s not found", econfig.Module)
	}
	return callback.NotifyConfigCallback(ctx, econfig)
}

// NotifyRemoteModuleConfig notify remote process 's module config changed
func NotifyRemoteModuleConfig(ctx context.Context, econfig *NotifyModuleConfig) error {
	notifyAddress, ex := getProcessModuleConfigNotifyAddress(string(econfig.Process))
	if !ex {
		return errors.Errorf("module %s, process %s, config notify address is not found.", econfig.Module, econfig.Process)
	}
	ctxlog := log.WithContext(ctx).WithFields(log.Fields{
		"module":         econfig.Module,
		"process":        econfig.Process,
		"notify address": notifyAddress.NotifyAddress,
	})
	ctxlog.Infof("notify module config")

	admin := agent.NewAdmin(agent.DefaultAdminConf())
	client, err := admin.NewClient(notifyAddress.Process)
	if err != nil {
		ctxlog.Errorf("new client err:%+v", err)
		return err
	}

	resp := new(agent.AgentctlResponse)
	err = client.Call(notifyAddress.NotifyAddress, econfig, resp)
	if err != nil {
		ctxlog.Errorf("notify config err:%+v", err)
		return err
	}
	return nil
}

func NotifyAllModules(ctx context.Context) error {
	return notifyModules(ctx, nil, true)
}

func NotifyModules(ctx context.Context, modules []string) error {
	return notifyModules(ctx, modules, false)
}

func notifyModules(ctx context.Context, modules []string, all bool) error {
	moduleConfigs := GetModuleConfigs()
	modulesToNotify := make([]*NotifyModuleConfig, 0, len(moduleConfigs))
	for _, moduleConfTpl := range moduleConfigs {
		if moduleConfTpl.Disabled {
			log.WithContext(ctx).Infof("module %s config is disabled", moduleConfTpl.Module)
			continue
		}
		affected := false
		if all {
			affected = true
		} else {
			for _, module := range modules {
				if moduleConfTpl.Module == module {
					affected = true
					break
				}
			}
		}
		if !affected {
			continue
		}

		process := moduleConfTpl.Process
		moduleConf, err := GetFinalModuleConfig(moduleConfTpl.Module)
		if err != nil {
			log.WithContext(ctx).Error(err)
			return err
		}
		modulesToNotify = append(modulesToNotify, &NotifyModuleConfig{
			Process: process,
			Module:  moduleConfTpl.Module,
			Config:  moduleConf.Config,
		})
	}
	err := NotifyModuleConfigs(ctx, &VerifyConfigResult{
		ConfigVersion:  nil,
		UpdatedConfigs: modulesToNotify,
	})
	if err != nil {
		log.WithContext(ctx).Errorf("notify module configs %+v, err:%+v", modulesToNotify, err)
	}
	return err
}

func NotifyModuleConfigForHttp(ctx context.Context, nconfig *NotifyModuleConfig) error {
	ctxlog := log.WithContext(ctx)
	err := ReloadConfigFromFiles(ctx)
	if err != nil {
		return errors.Errorf("reload config err:%s", err)
	}

	moduleConf, err := GetFinalModuleConfig(nconfig.Module)
	if err != nil {
		return errors.Errorf("get module config err:%s", err)
	}

	if nconfig.Process != moduleConf.Process {
		ctxlog.Warnf("module %s process should be %s, not %s", nconfig.Module, moduleConf.Process, nconfig.Process)
	}

	if moduleConf.Disabled {
		return errors.Errorf("module %s config is disabled.", nconfig.Module)
	}

	if moduleConf.Process != string(CurProcess) {
		return errors.Errorf("process %s is not reached, cur process is %s", moduleConf.Process, CurProcess)
	}

	nconfig.Config = moduleConf.Config
	err = NotifyLocalModuleConfig(ctx, nconfig)
	if err != nil {
		return errors.Errorf("notify module config err:%s", err)
	}

	return nil
}
