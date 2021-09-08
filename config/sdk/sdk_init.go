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

package sdk

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/lib/crypto"
)

var (
	configPropertiesMetaOnce sync.Once
)

//SDKConfig SDK configuration
type SDKConfig struct {
	ConfigPropertiesDir string
	ModuleConfigDir     string
	CryptoPath          string              `yaml:"cryptoPath"`
	CryptoMethod        crypto.CryptoMethod `yaml:"cryptoMethod"`
}

//InitSDK SDK initialization:
//1. Load key-value meta information: only value is for users to change, and other content is put into the SDK.
//2. Initialize the key-value configuration and the configuration template of the business module.
//3. To load the business callback module, the business needs to provide a callback function, and the configuration can be obtained when the business is initialized or changed.
//4. The callback service of the configuration module may be cross-process, and the cross-process API can be configured. It itself serves as part of configuration management.
func InitSDK(conf config.SDKConfig) error {
	log.Infof("init sdk conf %+v", conf)

	if err := config.InitCrypto(conf.CryptoPath, conf.CryptoMethod); err != nil {
		log.Error(err)
		return err
	}

	// init configs
	err := initConfigs(conf.ConfigPropertiesDir, conf.ModuleConfigDir)
	if err != nil {
		log.Error(err)
		return err
	}

	// init config callbacks
	if err := RegisterConfigCallbacks(); err != nil {
		log.Error(err)
		return err
	}

	// set process notify address
	err = config.InitModuleTypeConfig(context.Background(), config.NotifyProcessConfigModuleType)
	if err != nil {
		log.Error(err)
	}
	return err
}

//initConfigs Initialize the key-value configuration and the configuration template of the business module.
func initConfigs(configPropertiesDir, moduleConfigDir string) error {
	err := config.InitConfigProperties(configPropertiesDir)
	if err != nil {
		log.Error(err)
		return err
	}
	err = config.InitModuleConfigs(moduleConfigDir)
	if err != nil {
		log.Error(err)
	}
	return err
}
