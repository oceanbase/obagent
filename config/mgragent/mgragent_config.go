/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package mgragent

import (
	"bytes"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/lib/crypto"
)

// Config for ob_mgragent process
type ManagerAgentConfig struct {
	Server       ServerConfig         `yaml:"server"`
	SDKConfig    config.SDKConfig     `yaml:"sdkConfig"`
	CryptoMethod crypto.CryptoMethod  `yaml:"cryptoMethod"`
	Install      config.InstallConfig `yaml:"install"`
	ShellfConfig config.ShellfConfig  `yaml:"shellf"`
}

type AgentProxyConfig struct {
	ProxyAddress string `yaml:"proxyAddress"`
	ProxyEnabled bool   `yaml:"proxyEnabled"`
}

type ServerConfig struct {
	//Port    int    `yaml:"port"`
	Address string `yaml:"address"`
	RunDir  string `yaml:"runDir"`
}

func NewManagerAgentConfig(configFile string) *ManagerAgentConfig {
	_, err := os.Stat(configFile)
	if err != nil {
		log.WithField("file", configFile).WithError(err).Fatal("config file not found")
	}

	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.WithField("file", configFile).WithError(err).Fatal("fail to read config file")
	}

	config := new(ManagerAgentConfig)
	err = yaml.NewDecoder(bytes.NewReader(content)).Decode(config)
	if err != nil {
		log.WithField("file", configFile).WithError(err).Fatal("fail to decode config file")
	}
	return config
}
