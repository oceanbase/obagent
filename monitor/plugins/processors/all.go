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

package processors

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/processors/aggregate"
	"github.com/oceanbase/obagent/monitor/plugins/processors/attr"
	"github.com/oceanbase/obagent/monitor/plugins/processors/jointable"
	"github.com/oceanbase/obagent/monitor/plugins/processors/retag"
	"github.com/oceanbase/obagent/monitor/plugins/processors/slsmetric"
	"github.com/oceanbase/obagent/monitor/plugins/processors/transformer"
)

func init() {

	plugins.GetProcessorManager().Register("retagProcessor", func(config *monagent.PluginConfig) (plugins.Processor, error) {
		retagProcessor := &retag.RetagProcessor{}
		err := retagProcessor.Init(context.Background(), config.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("init retagProcessor failed")
			return nil, err
		}
		return retagProcessor, nil
	})
	plugins.GetProcessorManager().Register("aggregateProcessor", func(conf *monagent.PluginConfig) (plugins.Processor, error) {
		config := conf.PluginInnerConfig
		aggregateProcessor := &aggregate.Aggregator{}
		err := aggregateProcessor.Init(context.Background(), config)
		if err != nil {
			log.WithError(err).Error("init aggregateProcessor failed")
			return nil, err
		}
		return aggregateProcessor, nil
	})
	plugins.GetProcessorManager().Register("attrProcessor", func(conf *monagent.PluginConfig) (plugins.Processor, error) {
		config := conf.PluginInnerConfig
		attrProcessor := &attr.AttrProcessor{}
		err := attrProcessor.Init(context.Background(), config)
		if err != nil {
			log.WithError(err).Error("init attrProcessor failed")
			return nil, err
		}
		return attrProcessor, nil
	})
	plugins.GetProcessorManager().Register("jointable", func(conf *monagent.PluginConfig) (plugins.Processor, error) {
		config := conf.PluginInnerConfig
		processor := &jointable.JoinTable{}
		err := processor.Init(context.Background(), config)
		if err != nil {
			log.WithError(err).Error("init jointable processor failed")
			return nil, err
		}
		return processor, nil
	})
	plugins.GetProcessorManager().Register("logTransformer", func(config *monagent.PluginConfig) (plugins.Processor, error) {
		logTransformer := &transformer.LogTransformer{}
		err := logTransformer.Init(context.Background(), nil)
		if err != nil {
			log.WithError(err).Error("init logTransformer failed")
			return nil, err
		}
		return logTransformer, nil
	})
	plugins.GetProcessorManager().Register("slsmetric", func(conf *monagent.PluginConfig) (plugins.Processor, error) {
		slsMetricProcessor := slsmetric.NewSlsMetricProcessor()
		return slsMetricProcessor, nil
	})
}
