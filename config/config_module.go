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

package config

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	ManagerAgentBasicAuthConfigModuleType ModuleType = "mgragent.basic.auth"
	MonitorPipelineModuleType             ModuleType = "monagent.pipeline"
	MonitorServerBasicAuthModuleType      ModuleType = "monagent.server.basic.auth"
	ProxyConfigModuleType                 ModuleType = "proxy.config"
	NotifyProcessConfigModuleType         ModuleType = "module.config.notify"
	OBLogcleanerModuleType                ModuleType = "ob.logcleaner"
	MonitorLogConfigModuleType            ModuleType = "monagent.log.config"
	ManagerLogConfigModuleType            ModuleType = "mgragent.log.config"
	ManagerLogQuerierModuleType           ModuleType = "mgragent.logquerier"
	ConfigMetaModuleType                  ModuleType = "config.meta"
	StatConfigModuleType                  ModuleType = "stat.config"
)

const (
	ManagerAgentBasicAuthConfigModule = "mgragent.basic.auth"
	ManagerAgentProxyConfigModule     = "mgragent.proxy.config"
	NotifyProcessConfigModule         = "module.config.notify"
	OBLogcleanerModule                = "ob.logcleaner"
	MonitorLogConfigModule            = "monagent.log.config"
	ManagerLogConfigModule            = "mgragent.log.config"
	ManagerLogQuerierModule           = "mgragent.logquerier"
	ManagerAgentConfigMetaModule      = "mgragent.config.meta"
	MonitorAgentConfigMetaModule      = "monagent.config.meta"
)

var (
	mainModuleConfig *moduleConfigMain
)

type moduleConfigMain struct {
	moduleConfigGroups []*ModuleConfigGroup
	allModuleConfigs   map[string]ModuleConfig
	moduleConfigDir    string
}

type ModuleConfigGroup struct {
	Modules    []ModuleConfig `yaml:"modules"`
	ConfigFile string         `json:"-" yaml:"-"`
}

type ModuleConfig struct {
	Module     string      `yaml:"module"`
	ModuleType ModuleType  `yaml:"moduleType"`
	Disabled   bool        `yaml:"disabled"`
	Process    string      `yaml:"process"`
	Config     interface{} `yaml:"config"`
}

func GetModuleConfigs() map[string]ModuleConfig {
	return mainModuleConfig.allModuleConfigs
}

func GetFinalModuleConfig(module string) (*ModuleConfig, error) {
	sample, ex := GetModuleConfigs()[module]
	if !ex {
		return nil, errors.Errorf("module %s config is not found", module)
	}
	finalConf, err := getFinalModuleConfig(module, sample.Config, nil)
	if err != nil {
		return nil, errors.Errorf("get module %s config err:%s", module, err)
	}
	sample.Config = finalConf
	return &sample, nil
}

func getFinalModuleConfig(module string, sampleConf interface{}, diffkvs map[string]interface{}) (interface{}, error) {
	allkvs := GetConfigPropertiesKeyValues()
	for k, v := range diffkvs {
		allkvs[k] = fmt.Sprintf("%+v", v)
	}
	expander := NewExpanderWithKeyValues(DefaultExpanderPrefix, DefaultExpanderSuffix, allkvs, EnvironKeyValues())

	newConf, err := createModuleTypeInstance(module)
	if err != nil {
		return nil, err
	}

	finalConf, err := ToStructured(sampleConf, newConf, expander.Replace)
	return finalConf, err
}
