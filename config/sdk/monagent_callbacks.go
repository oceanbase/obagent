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

package sdk

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/api/web"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/monagent"
	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/monitor/engine"
)

// RegisterMonagentCallbacks When the service callback module is loaded,
// the service needs to provide the callback function,
// and the configuration can be obtained when the service is initialized or changed.
func RegisterMonagentCallbacks(ctx context.Context) error {
	err := config.RegisterConfigCallback(
		config.MonitorLogConfigModuleType,
		func() interface{} {
			return agentlog.LoggerConfig{}
		},
		setLogger,
		setLogger,
	)
	if err != nil {
		return err
	}

	// register pipeline module
	err = config.RegisterConfigCallback(
		config.MonitorPipelineModuleType,
		func() interface{} {
			return &monagent.PipelineModule{}
		},
		// Initialize the configuration callback
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*monagent.PipelineModule)
			if !ok {
				return errors.Errorf("init module %s conf %s is not *config.PipelineModule", config.MonitorPipelineModuleType, reflect.TypeOf(moduleConf))
			}
			err := engine.InitPipelineModuleCallback(ctx, conf)
			return err
		},

		// Configure update callbacks
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*monagent.PipelineModule)
			if !ok {
				return errors.Errorf("update module %s conf %s is not *config.PipelineModule", config.MonitorPipelineModuleType, reflect.TypeOf(moduleConf))
			}
			err := engine.UpdatePipelineModuleCallback(ctx, conf)
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
		// Initialize the configuration callback
		func(ctx context.Context, moduleConf interface{}) error {
			basicConf, ok := moduleConf.(config.BasicAuthConfig)
			if !ok {
				return errors.Errorf("init module %s conf %s is not config.BasicAuthConfig", config.MonitorServerBasicAuthModuleType, reflect.TypeOf(moduleConf))
			}
			notifyServerBasicAuth(basicConf)
			log.WithContext(ctx).Infof("module %s init config successfully", config.MonitorServerBasicAuthModuleType)
			return nil
		},
		// Configure update callbacks
		func(ctx context.Context, moduleConf interface{}) error {
			basicConf, ok := moduleConf.(config.BasicAuthConfig)
			if !ok {
				return errors.Errorf("update module %s conf %s is not config.BasicAuthConfig", config.MonitorServerBasicAuthModuleType, reflect.TypeOf(moduleConf))
			}
			notifyServerBasicAuth(basicConf)
			log.WithContext(ctx).Infof("module %s update config successfully", config.MonitorServerBasicAuthModuleType)
			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func notifyServerBasicAuth(basicConf config.BasicAuthConfig) error {
	monagentServer := web.GetMonitorAgentServer()
	monagentServer.Server.BasicAuthorizer.SetConf(basicConf)
	monagentServer.Server.UseBasicAuth()
	return nil
}
