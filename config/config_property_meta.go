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
	log "github.com/sirupsen/logrus"
)

var configPropertyMetas map[string]*ConfigProperty

func init() {
	configPropertyMetas = make(map[string]*ConfigProperty, 64)
}

func SetConfigPropertyMeta(property *ConfigProperty) {
	_, ex := configPropertyMetas[property.Key]
	if ex {
		log.Fatalf("config key %s already exist.", property.Key)
	}
	configPropertyMetas[property.Key] = property
}

func mergeConfigProperties() {
	for key, meta := range configPropertyMetas {
		property, ex := mainConfigProperties.allConfigProperties[key]
		if !ex {
			mainConfigProperties.allConfigProperties[key] = meta
			continue
		}
		if property.Value == meta.DefaultValue {
			if meta.Masked {
				// 数据需要脱敏，不打印内容
				log.Warnf("config key %s is still set as default value", key)
			} else {
				log.Warnf("config key %s is still set as default value:%+v", key, property.DefaultValue)
			}
		}
		log.Debugf("merge config meta and configfile config, config key %s", key)
		property.DefaultValue = meta.DefaultValue
		property.ValueType = meta.ValueType
		property.Encrypted = meta.Encrypted
		property.Fatal = meta.Fatal
		property.Masked = meta.Masked
		property.NeedRestart = meta.NeedRestart
		property.Description = meta.Description
		property.Unit = meta.Unit
		property.Valid = meta.Valid
	}
}
