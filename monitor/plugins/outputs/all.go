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

package outputs

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/outputs/es"
	"github.com/oceanbase/obagent/monitor/plugins/outputs/pushhttp"
)

func init() {
	plugins.GetOutputManager().Register("httpOutput", func(conf *monagent.PluginConfig) (plugins.Sink, error) {
		httpOutput := &pushhttp.HttpOutput{}
		err := httpOutput.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("init httpOutput failed")
			return nil, err
		}
		return httpOutput, nil
	})
	plugins.GetOutputManager().Register("esOutput", func(conf *monagent.PluginConfig) (plugins.Sink, error) {
		esOutput, err := es.NewESOutput(conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("NewESOutput failed")
			return nil, err
		}
		return esOutput, nil
	})

}
