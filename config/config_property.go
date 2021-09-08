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
	"fmt"
	"os"
	"reflect"
	"sort"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"
)

var (
	mainConfigProperties *ConfigPropertiesMain
)

const (
	maskedString = "****"
)

type ConfigPropertiesMain struct {
	ConfigGroups        []*ConfigPropertiesGroup
	allConfigProperties map[string]*ConfigProperty
	needRestartModules  map[string]*RestartModuleKeyValues
	configPropertiesDir string
}

// ContainsRestartNeededKeyValues Whether to include the configuration that needs to be restarted to take effect
func (c *ConfigPropertiesMain) ContainsRestartNeededKeyValues(module string, process string, kvs map[string]interface{}) bool {
	restatNeeded := false
	for key, value := range kvs {
		property, ex := c.allConfigProperties[key]
		if !ex {
			log.Errorf("config key %s is not found", key)
			continue
		}
		if !property.NeedRestart {
			continue
		}
		if _, ex := c.needRestartModules[module]; !ex {
			c.needRestartModules[module] = &RestartModuleKeyValues{
				Module:           module,
				Process:          process,
				RestartKeyValues: map[string]interface{}{},
			}
			continue
		}
		c.needRestartModules[module].RestartKeyValues[key] = value
		masedKeyValues := MaskedKeyValues(c.needRestartModules[module].RestartKeyValues)
		c.needRestartModules[module].RestartKeyValues = masedKeyValues
		log.Warnf("config key %s is changed, need restart agent process", key)
		restatNeeded = true
	}
	return restatNeeded
}

func (c *ConfigPropertiesMain) MaskedKeyValues(kvs map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{}, len(kvs))
	for key, value := range kvs {
		property, ex := c.allConfigProperties[key]
		if !ex {
			continue
		}
		if property.Masked {
			ret[key] = maskedString
		} else {
			ret[key] = value
		}
	}
	return ret
}

func (c *ConfigPropertiesMain) GetConfigPropertiesKeyValues() map[string]interface{} {
	ret := make(map[string]interface{}, len(c.allConfigProperties))
	for key, property := range c.allConfigProperties {
		ret[key] = property.Val()
	}
	return ret
}

func (c *ConfigPropertiesMain) GetDiffConfigProperties(keyValues map[string]interface{}, fatal bool) (map[string]interface{}, error) {
	properties := make(map[string]interface{}, 10)
	for key, value := range keyValues {
		property, ex := c.allConfigProperties[key]
		if !ex {
			return nil, errors.Errorf("key %s is not found in config properties!", key)
		}
		if !fatal && property.Fatal {
			return nil, errors.Errorf("key %s is fatal, cannot be changed by normal user.", key)
		}
		val, err := property.Parse(value)
		if err != nil {
			return nil, errors.Errorf("pase key %s failed, err:%+v", key, err)
		}
		// diff configs
		if property.Value != val {
			properties[key] = val
		}
	}

	return properties, nil
}

func (c *ConfigPropertiesMain) GetCurrentConfigVersion() *ConfigVersion {
	versions := []string{}
	dup := map[string]bool{}
	for _, group := range c.ConfigGroups {
		if dup[group.ConfigVersion] {
			continue
		}
		dup[group.ConfigVersion] = true
		versions = append(versions, group.ConfigVersion)
	}
	if len(versions) <= 0 {
		return &ConfigVersion{}
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})
	return &ConfigVersion{
		ConfigVersion: versions[0],
	}
}

func (c *ConfigPropertiesMain) saveIncrementalConfig(kvs map[string]interface{}) (*ConfigVersion, error) {
	groups := map[*ConfigPropertiesGroup]bool{}
	changed := false
	var err error
	for key, val := range kvs {
		property, ex := c.allConfigProperties[key]
		if !ex {
			log.WithField("config key", key).Errorf("config key not exist")
			continue
		}
		if property.Value != val {
			changed = true
		}
		finalVal := val
		if property.Encrypted {
			log.Debugf("encrypt config key %s", property.Key)
			rawVal := cast.ToString(val)
			finalVal, err = configCrypto.Encrypt(rawVal)
			if err != nil {
				return nil, errors.Errorf("Encrypt config key %s err:%s", property.Key, err)
			}
		}
		c.allConfigProperties[key].Value = finalVal

		for _, group := range c.ConfigGroups {
			for _, property := range group.Configs {
				if property.Key == key {
					property.Value = finalVal
					groups[group] = true
				}
			}
		}
	}
	if !changed {
		return nil, errors.Errorf("all configs is not changed")
	}

	var reterr error
	configVersion := generateNewConfigVersion()
	for group := range groups {
		group.ConfigVersion = configVersion.ConfigVersion
		err := group.SaveConfig()
		if err != nil {
			log.WithField("ConfigGroup", group).Errorf("save config to file err:%+v", err)
			reterr = err
		} else {
			log.Infof("save config %s to file %s", group.ConfigVersion, group.ConfigFile)
		}
	}
	err = snapshotForConfigVersion(configVersion.ConfigVersion)
	if err != nil {
		log.Errorf("save config snapshot %s err:%+v", configVersion.ConfigVersion, err)
		reterr = err
	}

	return configVersion, reterr
}

func (c *ConfigPropertiesMain) addConfigs(configRegion *ConfigPropertiesGroup) error {
	c.ConfigGroups = append(c.ConfigGroups, configRegion)
	for _, property := range configRegion.Configs {
		key := property.Key
		if _, ex := c.allConfigProperties[key]; ex {
			return errors.Errorf("key %s exists", key)
		}
		c.allConfigProperties[key] = property
	}
	return nil
}

type ConfigPropertiesGroup struct {
	ConfigVersion string            `json:"configVersion" yaml:"configVersion"`
	Configs       []*ConfigProperty `json:"configs" yaml:"configs"`
	ConfigFile    string            `json:"-" yaml:"-"`
}

func (g *ConfigPropertiesGroup) SaveConfig() error {
	file, err := os.OpenFile(g.ConfigFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return yaml.NewEncoder(file).Encode(g)
}

type ConfigProperty struct {
	Key          string                        `json:"key" yaml:"key"`
	Value        interface{}                   `json:"value" yaml:"value"`
	DefaultValue interface{}                   `json:"-" yaml:"-"`
	ValueType    ValueType                     `json:"valueType" yaml:"valueType"`
	Encrypted    bool                          `json:"encrypted" yaml:"encrypted"`
	Fatal        bool                          `json:"-" yaml:"-"`
	Masked       bool                          `json:"-" yaml:"-"`
	Description  string                        `json:"-" yaml:"-"`
	Unit         string                        `json:"-" yaml:"-"`
	NeedRestart  bool                          `json:"-" yaml:"-"`
	Valid        func(value interface{}) error `json:"-" yaml:"-"`
}

func (c *ConfigProperty) Val() interface{} {
	var val interface{}
	val = c.Value
	if c.Encrypted {
		rawVal := cast.ToString(val)
		// val is nil, no need to decrypt
		if len(rawVal) == 0 {
			return rawVal
		}
		log.Debugf("decrypt config key %s", c.Key)
		finalVal, err := configCrypto.Decrypt(rawVal)
		if err != nil {
			log.Errorf("Decrypt config key %s, value %s err:%+v", c.Key, c.Value, err)
		}
		return finalVal
	}
	return val
}

func (c *ConfigProperty) Parse(value interface{}) (val interface{}, err error) {
	defer func() {
		if err != nil || c.Valid == nil {
			return
		}
		if validateErr := c.Valid(val); validateErr != nil {
			err = errors.Errorf("validate %+v failed, err:%s", val, validateErr)
		}
	}()

	switch c.ValueType {
	case ValueBool:
		val, err = cast.ToBoolE(value)
		if err != nil {
			return nil, errors.Errorf("assert %s %+v (%s) to bool, err:%+v", c.Key, value, reflect.TypeOf(value), err)
		}
		return val, nil
	case ValueInt64:
		val, err = cast.ToInt64E(value)
		if err != nil {
			return nil, errors.Errorf("assert %s %+v (%s) to int64, err:%+v", c.Key, value, reflect.TypeOf(value), err)
		}
		return val, nil
	case ValueFloat64:
		val, err = cast.ToFloat64E(value)
		if err != nil {
			return nil, errors.Errorf("assert %s %+v (%s) to numberic float64, err:%+v", c.Key, value, reflect.TypeOf(value), err)
		}
		return val, nil
	case ValueString:
		val, err = cast.ToStringE(value)
		if err != nil {
			return nil, errors.Errorf("assert %s %+v (%s) to string, err:%+v", c.Key, value, reflect.TypeOf(value), err)
		}
		return val, nil
	default:
		return nil, errors.Errorf("key %s unsurported valueType %s", c.Key, c.ValueType)
	}
}

type ValueType string

const (
	ValueBool    ValueType = "bool"
	ValueInt64   ValueType = "int64"
	ValueFloat64 ValueType = "float64"
	ValueString  ValueType = "string"
)

func MaskedKeyValues(kvs map[string]interface{}) map[string]interface{} {
	return mainConfigProperties.MaskedKeyValues(kvs)
}

func GetConfigPropertiesKeyValues() map[string]string {
	kvs := mainConfigProperties.GetConfigPropertiesKeyValues()
	ret := make(map[string]string, len(kvs))
	for key, val := range kvs {
		ret[key] = fmt.Sprintf("%+v", val)
	}
	return ret
}

func GetDiffConfigProperties(keyValues map[string]interface{}, fatal bool) (map[string]interface{}, error) {
	return mainConfigProperties.GetDiffConfigProperties(keyValues, fatal)
}

func GetCurrentConfigVersion() *ConfigVersion {
	return mainConfigProperties.GetCurrentConfigVersion()
}

type RestartModuleKeyValues struct {
	Module           string
	Process          string
	RestartKeyValues map[string]interface{}
}

func ContainsRestartNeededKeyValues(module string, process string, kvs map[string]interface{}) bool {
	return mainConfigProperties.ContainsRestartNeededKeyValues(module, process, kvs)
}
