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

package config

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/api/response"
)

var (
	processConfigNotifyAddresses     map[string]ProcessConfigNotifyAddress
	processConfigNotifyAddressesLock sync.Mutex
)

func init() {
	processConfigNotifyAddresses = make(map[string]ProcessConfigNotifyAddress, 4)
}

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
				errs = append(errs, errors.Errorf("init module %s err:%+v", m, err).Error())
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
	log.WithContext(ctx).Infof("notify moduel configs length:%d", len(verifyConfigResult.UpdatedConfigs))
	for _, conf := range verifyConfigResult.UpdatedConfigs {
		err := notifyModuleConfig(ctx, conf)
		if err != nil {
			return err
		}
	}
	return nil
}

// notify module config changed
func notifyModuleConfig(ctx context.Context, econfig *NotifyModuleConfig) error {
	log.Infof("module %s, notify process %s, current process %s", econfig.Module, econfig.Process, CurProcess)
	notifyAddress, ex := getProcessModuleConfigNotifyAddress(string(econfig.Process))
	if !ex {
		return errors.Errorf("process %s notify config is not found", econfig.Process)
	}
	if !notifyAddress.Local && econfig.Process != CurProcess {
		return NotifyRemoteModuleConfig(ctx, econfig)
	}
	callback, ex := getModuleCallback(econfig.Module)
	if !ex {
		return errors.Errorf("module %s not found", econfig.Module)
	}
	return callback.NotifyConfigCallback(ctx, econfig)
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
	data, err := json.Marshal(econfig)
	if err != nil {
		err = errors.Errorf("module %s, process %s, json marshal err:%+v", econfig.Module, econfig.Process, err)
		ctxlog.Errorf("json marshal err:%+v", err)
		return err
	}
	ctxlog.Infof("notify module config")

	client := &http.Client{Timeout: time.Second * 1}
	req, err := http.NewRequest("POST", notifyAddress.NotifyAddress, bytes.NewReader(data))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		ctxlog.Errorf("notify config new request err:%+v", err)
		return err
	}
	req.SetBasicAuth(notifyAddress.AuthConfig.Username, notifyAddress.AuthConfig.Password)
	resp, err := client.Do(req)
	if err != nil {
		ctxlog.Errorf("notify config do request err:%+v", err)
		return err
	}
	defer resp.Body.Close()

	r := new(response.AgentResponse)
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		ctxlog.Errorf("decode resp err:%+v", err)
		return err
	}

	if !r.Successful || r.Error != nil {
		return errors.Errorf("errorCode:%d, errorMessage:%s", r.Error.Code, r.Error.Message)
	}

	return nil
}

func NotifyModules(modules []string) error {
	moduleConfigs := GetModuleConfigs()
	notifyModules := make([]*NotifyModuleConfig, 0, len(modules))
	for _, moduleConf := range moduleConfigs {
		for _, module := range modules {
			if moduleConf.Module == module {
				process := moduleConf.Process
				moduleConf, err := GetFinalModuleConfig(module)
				if err != nil {
					log.Error(err)
					return err
				}
				if moduleConf.Disabled {
					log.Warnf("module %s config is disabled", module)
					continue
				}
				notifyModules = append(notifyModules, &NotifyModuleConfig{
					Process: Process(process),
					Module:  module,
					Config:  moduleConf.Config,
				})
			}
		}
	}
	err := NotifyModuleConfigs(context.Background(), &VerifyConfigResult{
		ConfigVersion:  nil,
		UpdatedConfigs: notifyModules,
	})
	if err != nil {
		log.Errorf("notify module configs %+v, err:%+v", notifyModules, err)
	}
	return err
}

func NotifyModuleConfigForHttp(ctx context.Context, nconfig *NotifyModuleConfig) error {
	ctxlog := log.WithContext(ctx)
	err := ReloadConfigFromFiles()
	if err != nil {
		return errors.Errorf("reload config err:%+v", err)
	}

	moduleConf, err := GetFinalModuleConfig(nconfig.Module)
	if err != nil {
		return errors.Errorf("get module config err:%+v", err)
	}

	if nconfig.Process != Process(moduleConf.Process) {
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
		return errors.Errorf("notify module config err:%+v", err)
	}

	return nil
}
