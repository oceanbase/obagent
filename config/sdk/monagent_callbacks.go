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
	"reflect"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/engine"
)

//RegisterMonagentCallbacks To load the business callback module,
//the business needs to provide a callback function,
//and the configuration can be obtained when the business is initialized or changed.
func RegisterMonagentCallbacks() error {
	// register pipeline module
	err := config.RegisterConfigCallback(
		config.MonitorPipelineModuleType,
		func() interface{} {
			return &config.PipelineModule{}
		},
		// Initial configuration callback
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*config.PipelineModule)
			if !ok {
				return errors.Errorf("init module %s conf %s is not *config.PipelineModule", config.MonitorPipelineModuleType, reflect.TypeOf(moduleConf))
			}
			err := engine.InitPipelineModuleCallback(conf)
			return err
		},

		// Configuration update callback
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*config.PipelineModule)
			if !ok {
				return errors.Errorf("update module %s conf %s is not *config.PipelineModule", config.MonitorPipelineModuleType, reflect.TypeOf(moduleConf))
			}
			err := engine.UpdatePipelineModuleCallback(conf)
			return err
		},
	)
	if err != nil {
		return err
	}

	// monagent server basic auth
	err = config.RegisterConfigCallback(
		config.MonitorServerBasicAuthModuleType,
		func() interface{} {
			return config.BasicAuthConfig{}
		},
		// Initial configuration callback
		func(ctx context.Context, moduleConf interface{}) error {
			basicConf, ok := moduleConf.(config.BasicAuthConfig)
			if !ok {
				return errors.Errorf("init module %s conf %s is not config.BasicAuthConfig", config.MonitorServerBasicAuthModuleType, reflect.TypeOf(moduleConf))
			}
			engine.NotifyServerBasicAuth(basicConf)
			log.WithContext(ctx).Infof("module %s init config sucessfully", config.MonitorServerBasicAuthModuleType)
			return nil
		},
		// Configuration update callback
		func(ctx context.Context, moduleConf interface{}) error {
			basicConf, ok := moduleConf.(config.BasicAuthConfig)
			if !ok {
				return errors.Errorf("update module %s conf %s is not config.BasicAuthConfig", config.MonitorServerBasicAuthModuleType, reflect.TypeOf(moduleConf))
			}
			engine.NotifyServerBasicAuth(basicConf)
			log.WithContext(ctx).Infof("module %s update config sucessfully", config.MonitorServerBasicAuthModuleType)
			return nil
		},
	)
	if err != nil {
		return err
	}

	// monagent admin server basic auth
	err = config.RegisterConfigCallback(
		config.MonitorAdminBasicAuthModuleType,
		func() interface{} {
			return config.BasicAuthConfig{}
		},
		// Initial configuration callback
		func(ctx context.Context, moduleConf interface{}) error {
			basicConf, ok := moduleConf.(config.BasicAuthConfig)
			if !ok {
				return errors.Errorf("init module %s conf %s is not config.BasicAuthConfig", config.MonitorAdminBasicAuthModuleType, reflect.TypeOf(moduleConf))
			}
			engine.NotifyAdminServerBasicAuth(basicConf)
			log.WithContext(ctx).Infof("module %s init config sucessfully", config.MonitorAdminBasicAuthModuleType)
			return nil
		},
		// Configuration update callback
		func(ctx context.Context, moduleConf interface{}) error {
			basicConf, ok := moduleConf.(config.BasicAuthConfig)
			if !ok {
				return errors.Errorf("update module %s conf %s is not config.BasicAuthConfig", config.MonitorAdminBasicAuthModuleType, reflect.TypeOf(moduleConf))
			}
			engine.NotifyAdminServerBasicAuth(basicConf)
			log.WithContext(ctx).Infof("module %s update config sucessfully", config.MonitorAdminBasicAuthModuleType)
			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}
