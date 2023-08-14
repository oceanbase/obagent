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

package inputs

import (
	"context"
	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/inputs/host"
	"github.com/oceanbase/obagent/monitor/plugins/inputs/net"
	"github.com/oceanbase/obagent/monitor/plugins/inputs/obcommon"
	"github.com/oceanbase/obagent/monitor/plugins/inputs/process"
	log "github.com/sirupsen/logrus"
)

func init() {
	plugins.GetInputManager().Register("netConnectivityInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		connectivityInput := &net.ConnectivityInput{}
		err := connectivityInput.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Errorf("init connectivityInput failed")
			return nil, err
		}
		return connectivityInput, nil
	})
	plugins.GetInputManager().Register("dbConnectivityInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		connectivityInput := &obcommon.ConnectivityInput{}
		err := connectivityInput.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Errorf("init connectivityInput failed")
			return nil, err
		}
		return connectivityInput, nil
	})
	plugins.GetInputManager().Register("processInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		processInput := &process.ProcessInput{}
		err := processInput.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Errorf("init precessInput failed")
			return nil, err
		}
		return processInput, nil
	})
	plugins.GetInputManager().Register("hostCustomInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		customInput := &host.CustomInput{}
		err := customInput.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Errorf("init customInput failed")
			return nil, err
		}
		return customInput, nil
	})
}
