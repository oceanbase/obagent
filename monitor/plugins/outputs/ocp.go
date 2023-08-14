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

package outputs

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/outputs/sls"
)

func init() {
	plugins.GetOutputManager().Register("slsOutput", func(conf *monagent.PluginConfig) (plugins.Sink, error) {
		configBytes, err := yaml.Marshal(conf.PluginInnerConfig)
		if err != nil {
			log.Errorf("marshal slsOutput config failed, err: %s", err)
			return nil, err
		}
		var slsOutputConfig = &sls.Config{}
		err = yaml.Unmarshal(configBytes, slsOutputConfig)
		if err != nil {
			log.Errorf("unmarshal slsOutput config failed, err: %s", err)
			return nil, err
		}
		slsOutput := sls.NewSLSOutput(slsOutputConfig)
		return slsOutput, nil
	})
}
