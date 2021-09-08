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
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ConfigVersion Configuration version number
type ConfigVersion struct {
	ConfigVersion string
}

func generateNewConfigVersion() *ConfigVersion {
	return &ConfigVersion{
		ConfigVersion: time.Now().Format("2006-01-02T15:04:05.9999Z07:00"),
	}
}

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type KeyValues struct {
	Configs []KeyValue `json:"configs"`
}

type VerifyConfigResult struct {
	ConfigVersion  *ConfigVersion
	UpdatedConfigs []*NotifyModuleConfig
}

type NotifyModuleConfig struct {
	Process          Process                `json:"process"`
	Module           string                 `json:"module"`
	Config           interface{}            `json:"-"`
	UpdatedKeyValues map[string]interface{} `json:"updatedKeyValues"`
}

func UpdateConfig(ctx context.Context, kvs *KeyValues, async bool) (*ConfigVersion, error) {
	keyValues := make(map[string]interface{})
	for _, kv := range kvs.Configs {
		keyValues[kv.Key] = kv.Value
	}
	ctxlog := log.WithContext(ctx)

	ctxlog.WithField("config key values", MaskedKeyValues(keyValues)).Info("update module config")

	result, err := verifyAndSaveConfig(ctx, keyValues)
	if err != nil {
		log.WithContext(ctx).Errorf("VerifyAndSaveConfig err:%+v", err)
		return nil, err
	}

	notifyFunc := func() error {
		err := NotifyModuleConfigs(ctx, result)
		if err != nil {
			ctxlog.Errorf("update module config failed, config version:%s, changed modules length:%d, err:%+v",
				result.ConfigVersion.ConfigVersion, len(result.UpdatedConfigs), err)
		}
		return err
	}
	if async {
		go notifyFunc()
	} else {
		err := notifyFunc()
		return nil, err
	}

	return result.ConfigVersion, nil
}

func UpdateConfigPairs(pairs []string) error {
	log.WithField("pairs", pairs).Debugf("update module configs")
	kvs := make([]KeyValue, 0, len(pairs))
	for _, pair := range pairs {
		key, value, err := parseKeyValue(pair)
		if err != nil {
			log.Errorf("parse pair %s, err:%+v", pair, err)
			return err
		}
		kvs = append(kvs, KeyValue{
			Key:   key,
			Value: value,
		})
	}
	log.WithField("key-values", kvs).Debugf("update module configs")

	configVersion, err := UpdateConfig(context.Background(), &KeyValues{Configs: kvs}, true)
	if err != nil {
		log.Errorf("update config, err:%+v", err)
		return err
	}
	log.Infof("update config success, version:%+v", configVersion)
	return nil
}

func parseKeyValue(pairStr string) (string, string, error) {
	pair := strings.Split(pairStr, "=")
	if len(pair) != 2 {
		return "", "", errors.Errorf("key-value pair %s is not a valid key=value formatted.", pairStr)
	}
	return pair[0], pair[1], nil
}

func verifyConfigs(ctx context.Context, keyValues map[string]interface{}) (map[string]interface{}, error) {
	maskedKeyValues := MaskedKeyValues(keyValues)
	ctxlog := log.WithContext(ctx)
	ctxlog.WithField("key-values", maskedKeyValues).Info("update config")

	diffkvs, err := GetDiffConfigProperties(keyValues, false)
	if err != nil {
		ctxlog.WithField("key-values", maskedKeyValues).Errorf("GetDiffConfigProperties err:%+v", err)
		return nil, err
	}
	return diffkvs, nil
}

func verifyAndSaveConfig(ctx context.Context, keyValues map[string]interface{}) (*VerifyConfigResult, error) {
	ctxlog := log.WithContext(ctx)

	if len(keyValues) <= 0 {
		return nil, errors.Errorf("request keyValues is nil")
	}

	diffkvs, err := verifyConfigs(ctx, keyValues)
	if err != nil {
		return nil, err
	}
	if len(diffkvs) <= 0 {
		currentVersion := GetCurrentConfigVersion()
		log.Infof("configs do not change, no need to modify, current config version:%+v.", currentVersion)
		return &VerifyConfigResult{
			ConfigVersion: currentVersion,
		}, nil
	}

	updatedConfigs, err := getUpdatedConfigs(ctx, diffkvs)
	if err != nil {
		ctxlog.WithFields(log.Fields{
			"updated configs": MaskedKeyValues(diffkvs),
		}).Errorf("getUpdatedConfigs err:%+v", err)
		return nil, err
	}

	configVersion, err := saveIncrementalConfig(diffkvs)
	if err != nil {
		ctxlog.WithFields(log.Fields{
			"configVersion":   configVersion,
			"updated configs": MaskedKeyValues(diffkvs),
		}).Errorf("save config failed, please try again later or contant administrator. err:%+v", err)
		return nil, err
	}

	return &VerifyConfigResult{
		ConfigVersion:  configVersion,
		UpdatedConfigs: updatedConfigs,
	}, nil
}

func getUpdatedConfigs(ctx context.Context, diffkvs map[string]interface{}) ([]*NotifyModuleConfig, error) {
	ctxlog := log.WithContext(ctx)
	kvs := make(map[string]string, len(diffkvs))
	for k, v := range diffkvs {
		kvs[k] = fmt.Sprintf("%+v", v)
	}
	modules := GetModuleConfigs()
	expander := NewExpanderWithKeyValues(DefaultExpanderPrefix, DefaultExpanderSuffix, kvs, EnvironKeyValues())
	paths, err := ReplacedPath(modules, expander.Replace)
	if err != nil {
		return nil, err
	}
	ctxlog.WithField("config replaced path", paths).Debug("ReplacedPath")

	changedConfigs := make([]*NotifyModuleConfig, 0, 2)
	moduleKeys := make(map[string][]string, 2)
	for _, path := range paths {
		if len(path) <= 0 {
			continue
		}
		module := path[0]
		if _, ex := moduleKeys[module]; !ex {
			moduleKeys[module] = make([]string, 0, 4)
			moduleKeys[module] = append(moduleKeys[module], expander.TrimPrefixSuffix(path[len(path)-1])...)
		} else {
			moduleKeys[module] = append(moduleKeys[module], expander.TrimPrefixSuffix(path[len(path)-1])...)
			continue
		}
		sample := modules[module]
		finalConf, err := getFinalModuleConfig(module, sample.Config, diffkvs)
		if err != nil {
			return nil, err
		}
		changedConfig := &NotifyModuleConfig{
			Process: Process(sample.Process),
			Module:  module,
			Config:  finalConf,
		}
		changedConfigs = append(changedConfigs, changedConfig)
		ctxlog.Infof("changed config, module %s, process %s", changedConfig.Module, changedConfig.Process)
	}
	ctxlog.WithField("change config modules", moduleKeys).Info("changed module")
	if len(changedConfigs) <= 0 {
		return nil, errors.Errorf("no config changed")
	}
	for _, conf := range changedConfigs {
		moduleChangedKeys, ex := moduleKeys[conf.Module]
		if !ex {
			continue
		}
		moduleDiffkvs := make(map[string]interface{}, len(moduleChangedKeys))
		for _, key := range moduleChangedKeys {
			moduleDiffkvs[key] = diffkvs[key]
		}
		needRestart := ContainsRestartNeededKeyValues(conf.Module, string(conf.Process), moduleDiffkvs)
		if needRestart {
			ctxlog.Warnf("module %s changed config %+v contain restart keys, please restart agent later", conf.Module, MaskedKeyValues(moduleDiffkvs))
		}
		conf.UpdatedKeyValues = moduleDiffkvs
	}

	return changedConfigs, nil
}

func saveIncrementalConfig(kvs map[string]interface{}) (*ConfigVersion, error) {
	return mainConfigProperties.saveIncrementalConfig(kvs)
}

func snapshotForConfigVersion(configVersion string) error {
	snapshotPath := filepath.Join(filepath.Dir(mainConfigProperties.configPropertiesDir), configVersion)

	var errs error
	err := snapshotForPath(snapshotPath, mainConfigProperties.configPropertiesDir)
	if err != nil {
		log.Errorf("save config snapshot %s err:%+v", configVersion, err)
		errs = errors.Errorf("%s, err:%s", errs, err)
	}
	err = snapshotForPath(snapshotPath, mainModuleConfig.moduleConfigDir)
	if err != nil {
		log.Errorf("save module config snapshot %s to dir %s err:%+v", configVersion, snapshotPath, err)
		errs = errors.Errorf("%s, err:%s", errs, err)
	}
	return errs
}
