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

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/mgragent"
	"github.com/oceanbase/obagent/executor/cleaner"
	"github.com/oceanbase/obagent/executor/log_query"
	"github.com/oceanbase/obagent/lib/http"
	agentlog "github.com/oceanbase/obagent/log"
)

// RegisterMgragentCallbacks When the service callback module is loaded,
// the service needs to provide the callback function,
// and the configuration can be obtained when the service is initialized or changed.
func RegisterMgragentCallbacks(ctx context.Context) error {
	err := config.RegisterConfigCallback(
		config.ManagerLogConfigModuleType,
		func() interface{} {
			return agentlog.LoggerConfig{}
		},
		setLogger,
		setLogger,
	)
	if err != nil {
		return err
	}

	err = config.RegisterConfigCallback(
		config.ManagerAgentBasicAuthConfigModuleType,
		func() interface{} {
			return config.BasicAuthConfig{}
		},
		// Initialize the configuration callback
		func(ctx context.Context, moduleConf interface{}) error {
			basicConf, ok := moduleConf.(config.BasicAuthConfig)
			if !ok {
				return errors.Errorf("init module %s conf %s is not config.BasicAuthConfig", config.ManagerAgentBasicAuthConfigModuleType, reflect.TypeOf(moduleConf))
			}
			common.NotifyConf(basicConf)
			log.WithContext(ctx).Infof("module %s init config successfully", config.ManagerAgentBasicAuthConfigModuleType)
			return nil
		},
		// Configure update callbacks
		func(ctx context.Context, moduleConf interface{}) error {
			basicConf, ok := moduleConf.(config.BasicAuthConfig)
			if !ok {
				return errors.Errorf("update module %s conf %s is not config.BasicAuthConfig", config.ManagerAgentBasicAuthConfigModuleType, reflect.TypeOf(moduleConf))
			}
			common.NotifyConf(basicConf)
			log.WithContext(ctx).Infof("module %s update config successfully", config.ManagerAgentBasicAuthConfigModuleType)
			return nil
		},
	)
	if err != nil {
		return err
	}

	err = config.RegisterConfigCallback(
		config.OBLogcleanerModuleType,
		func() interface{} {
			return new(mgragent.ObCleanerConfig)
		},
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*mgragent.ObCleanerConfig)
			if !ok {
				return errors.Errorf("module %s init conf %s is not *config.ObCleanerConfig", config.OBLogcleanerModule, reflect.TypeOf(moduleConf))
			}
			err := cleaner.InitOBCleanerConf(conf)
			if err != nil {
				return errors.Errorf("init ob cleaner err:%s", err)
			}
			return nil
		},
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*mgragent.ObCleanerConfig)
			if !ok {
				return errors.Errorf("module %s update conf %s is not *config.ObCleanerConfig", config.OBLogcleanerModule, reflect.TypeOf(moduleConf))
			}
			err := cleaner.UpdateOBCleanerConf(conf)
			if err != nil {
				return errors.Errorf("update ob cleaner err:%s", err)
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	err = config.RegisterConfigCallback(
		config.ManagerLogQuerierModuleType,
		func() interface{} {
			return new(mgragent.LogQueryConfig)
		},
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*mgragent.LogQueryConfig)
			if !ok {
				return errors.Errorf("module %s init conf %s is not *config.LogQueryConfig", config.ManagerLogQuerierModuleType, reflect.TypeOf(moduleConf))
			}
			err := log_query.InitLogQuerierConf(conf)
			if err != nil {
				return errors.Errorf("init log querier conf err:%s", err)
			}
			return nil
		},
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*mgragent.LogQueryConfig)
			if !ok {
				return errors.Errorf("module %s update conf %s is not *config.LogQueryConfig", config.ManagerLogQuerierModuleType, reflect.TypeOf(moduleConf))
			}
			err := log_query.UpdateLogQuerierConf(conf)
			if err != nil {
				return errors.Errorf("update log querier conf err:%s", err)
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	err = config.RegisterConfigCallback(
		config.ProxyConfigModuleType,
		func() interface{} {
			return new(mgragent.AgentProxyConfig)
		},
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*mgragent.AgentProxyConfig)
			if !ok {
				return errors.Errorf("module %s init conf %s is not *mgragent.AgentProxyConfig", config.ProxyConfigModuleType, reflect.TypeOf(moduleConf))
			}
			if conf.ProxyEnabled {
				err := http.SetSocksProxy(conf.ProxyAddress)
				if err != nil {
					return errors.Errorf("SetSocksProxy err:%s", err)
				}
			}

			return nil
		},
		func(ctx context.Context, moduleConf interface{}) error {
			conf, ok := moduleConf.(*mgragent.AgentProxyConfig)
			if !ok {
				return errors.Errorf("module %s update conf %s is not *mgragent.AgentProxyConfig", config.ProxyConfigModuleType, reflect.TypeOf(moduleConf))
			}
			if conf.ProxyEnabled {
				err := http.SetSocksProxy(conf.ProxyAddress)
				if err != nil {
					return errors.Errorf("SetSocksProxy err:%s", err)
				}
			} else {
				http.UnsetSocksProxy()
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}
