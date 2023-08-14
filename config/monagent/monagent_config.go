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

package monagent

import (
	"bytes"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/lib/crypto"
)

type MonitorAgentConfig struct {
	// monagent server config
	Server MonitorAgentHttpConfig `yaml:"server"`
	// pipeline config file dir
	ModulePath string `yaml:"modulePath"`
	// pipeline config file dir
	PropertiesPath string `yaml:"propertiesPath"`
	// encrypt key path
	CryptoPath string `yaml:"cryptoPath"`
	// crypto method
	CryptoMethod crypto.CryptoMethod `yaml:"cryptoMethod"`
}

type MonitorAgentHttpConfig struct {
	// monitor metrics server address
	Address string `yaml:"address"`

	RunDir string `yaml:"runDir"`
}

// DecodeMonitorAgentServerConfig decode yaml formatted configfile, return MonitorAgentConfig
// it will be fatal if failed.
func DecodeMonitorAgentServerConfig(configfile string) (*MonitorAgentConfig, error) {
	_, err := os.Stat(configfile)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(configfile)
	if err != nil {
		return nil, err
	}

	config := new(MonitorAgentConfig)
	err = yaml.NewDecoder(bytes.NewReader(content)).Decode(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
