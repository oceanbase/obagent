package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/mask"
)

type ConfigVersion struct {
	ConfigVersion string
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

func UpdateConfig(ctx context.Context, kvs *KeyValues) (*ConfigVersion, error) {
	keyValues := make(map[string]interface{})
	for _, kv := range kvs.Configs {
		keyValues[kv.Key] = kv.Value
	}
	ctxlog := log.WithContext(ctx)

	ctxlog.WithField("config key values", MaskedKeyValues(keyValues)).Info("update module config")

	result, err := verifyAndSaveConfig(ctx, keyValues)
	if err != nil {
		log.WithContext(ctx).Errorf("VerifyAndSaveConfig err:%s", err)
		return nil, err
	}

	notifyErr := NotifyModuleConfigs(ctx, result)
	if notifyErr != nil {
		ctxlog.Errorf("update module config failed, config version:%s, changed modules length:%d, err:%+v",
			result.ConfigVersion.ConfigVersion, len(result.UpdatedConfigs), notifyErr)
	}

	return result.ConfigVersion, nil
}

func UpdateConfigPairs(ctx context.Context, pairs []string) error {
	maskedPairs := mask.MaskSlice(pairs)
	log.WithContext(ctx).WithField("pairs", maskedPairs).Info("update module configs")
	kvs, err := getKeyValues(pairs)
	if err != nil {
		log.WithContext(ctx).Errorf("getKeyValues, err:%+v", err)
		return err
	}

	configVersion, err := UpdateConfig(context.Background(), &KeyValues{Configs: kvs})
	if err != nil {
		log.WithContext(ctx).Errorf("update config, err:%+v", err)
		return err
	}
	log.WithContext(ctx).Infof("update config success, version:%+v", configVersion)
	return nil
}

func ValidateConfigPairs(ctx context.Context, pairs []string) error {
	maskedPairs := mask.MaskSlice(pairs)
	log.WithContext(ctx).WithField("pairs", maskedPairs).Info("validate module configs")
	kvs, err := getKeyValues(pairs)
	if err != nil {
		log.WithContext(ctx).Errorf("getKeyValues, err:%+v", err)
		return err
	}

	return ValidateConfigKeyValues(ctx, kvs)
}

func ValidateConfigKeyValues(ctx context.Context, kvs []KeyValue) error {
	log.WithContext(ctx).Info("validate module configs")

	configMain, err := decodeConfigPropertiesGroups(ctx, mainConfigProperties.configPropertiesDir)
	if err != nil {
		err = errors.Errorf("decode config properties from path %s, err:%s", mainConfigProperties.configPropertiesDir, err)
		log.WithContext(ctx).Error(err)
		return err
	}
	for _, kv := range kvs {
		property, ex := configMain.allConfigProperties[kv.Key]
		if !ex {
			return errors.Errorf("key %s is not exist in config.", kv.Key)
		}
		val := fmt.Sprintf("%+v", property.Val())
		if val != kv.Value {
			if !property.Masked {
				log.WithContext(ctx).Warnf("key %s is not equal with config, current value is %+v", kv.Key, val)
			}
			return errors.Errorf("key %s is not equal with config.", kv.Key)
		}
	}

	return nil
}

func getKeyValues(pairs []string) ([]KeyValue, error) {
	kvs := make([]KeyValue, 0, len(pairs))
	for _, pair := range pairs {
		key, value, err := parseKeyValue(pair)
		if err != nil {
			log.Errorf("parse pair %s, err:%+v", mask.Mask(pair), err)
			return nil, err
		}
		kvs = append(kvs, KeyValue{
			Key:   key,
			Value: value,
		})
	}

	return kvs, nil
}

func parseKeyValue(pairStr string) (string, string, error) {
	pair := strings.Split(pairStr, "=")
	if len(pair) != 2 {
		return "", "", errors.Errorf("key-value pair %s is not a valid key=value formatted.", mask.Mask(pairStr))
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
		ctxlog.Infof("configs do not change, no need to modify, current config version:%+v.", currentVersion)
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

	configVersion, err := saveIncrementalConfig(ctx, diffkvs)
	if err != nil {
		ctxlog.WithFields(log.Fields{
			"configVersion":   configVersion,
			"updated configs": MaskedKeyValues(diffkvs),
		}).Errorf("save config failed, please try again later or content administrator. err:%+v", err)
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
			Process: sample.Process,
			Module:  module,
			Config:  finalConf,
		}
		changedConfigs = append(changedConfigs, changedConfig)
		ctxlog.Infof("changed config, module %s, process %s", changedConfig.Module, changedConfig.Process)
	}
	ctxlog.WithField("change config modules", moduleKeys).Info("changed module")
	if len(changedConfigs) <= 0 {
		return changedConfigs, nil
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

func saveIncrementalConfig(ctx context.Context, kvs map[string]interface{}) (*ConfigVersion, error) {
	return mainConfigProperties.saveIncrementalConfig(ctx, kvs)
}

func snapshotForConfigVersion(ctx context.Context, configVersion string) error {
	snapshotPath := filepath.Join(filepath.Dir(mainConfigProperties.configPropertiesDir), configVersion)

	var errs error
	err := snapshotForPath(ctx, snapshotPath, mainConfigProperties.configPropertiesDir)
	if err != nil {
		log.WithContext(ctx).Errorf("save config snapshot %s err:%+v", configVersion, err)
		errs = errors.Errorf("%s, err:%s", errs, err)
	}
	err = snapshotForPath(ctx, snapshotPath, mainModuleConfig.moduleConfigDir)
	if err != nil {
		log.WithContext(ctx).Errorf("save module config snapshot %s to dir %s err:%+v", configVersion, snapshotPath, err)
		errs = errors.Errorf("%s, err:%s", errs, err)
	}
	configMetaBackupWorker.checkOnce(ctx)
	return errs
}
