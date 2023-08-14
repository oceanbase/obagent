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

package processors

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/processors/filter"
	"github.com/oceanbase/obagent/monitor/plugins/processors/label"
)

func init() {

	plugins.GetProcessorManager().Register("excludeFilterProcessor", func(conf *monagent.PluginConfig) (plugins.Processor, error) {
		config := conf.PluginInnerConfig
		excludeFilterProcessor := &filter.ExcludeFilter{}
		err := excludeFilterProcessor.Init(context.Background(), config)
		if err != nil {
			log.WithError(err).Error("init excludeFileterProcessor failed")
			return nil, err
		}
		return excludeFilterProcessor, nil
	})
	plugins.GetProcessorManager().Register("mountLabelProcessor", func(conf *monagent.PluginConfig) (plugins.Processor, error) {
		config := conf.PluginInnerConfig
		mountLabelProcessor := &label.MountLabelProcessor{}
		err := mountLabelProcessor.Init(context.Background(), config)
		if err != nil {
			log.WithError(err).Error("Init mountLabelProcessor failed")
			return nil, err
		}
		return mountLabelProcessor, nil
	})
}
