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

package inputs

import (
	"context"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/config/monagent"
	path2 "github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/inputs/log_tailer"
	"github.com/oceanbase/obagent/monitor/plugins/inputs/mysql"
	"github.com/oceanbase/obagent/monitor/plugins/inputs/nodeexporter"
	"github.com/oceanbase/obagent/monitor/plugins/inputs/prometheus"
)

func init() {
	plugins.GetInputManager().Register("mysqlTableInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		tableInput := &mysql.TableInput{}
		err := tableInput.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("init tableInput failed")
			return nil, err
		}
		return tableInput, nil
	})

	plugins.GetInputManager().Register("mysqldInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		mysqldInput := &mysql.MysqldInput{}
		err := mysqldInput.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("init mysqldExporter failed")
			return nil, err
		}
		return mysqldInput, nil
	})

	plugins.GetInputManager().Register("nodeExporterInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		nodeExporter := &nodeexporter.NodeExporter{}
		err := nodeExporter.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("init nodeExporter failed")
			return nil, err
		}
		return nodeExporter, nil
	})

	plugins.GetInputManager().Register("prometheusInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		config := conf.PluginInnerConfig
		prometheusInput := &prometheus.Prometheus{}
		err := prometheusInput.Init(context.Background(), config)
		if err != nil {
			log.WithError(err).Error("init prometheusInput failed")
			return nil, err
		}
		return prometheusInput, nil
	})
	plugins.GetInputManager().Register("logTailerInput", func(conf *monagent.PluginConfig) (plugins.Source, error) {
		configData, err := yaml.Marshal(conf.PluginInnerConfig)
		log.Infof("configData: %+v", *conf)
		if err != nil {
			log.WithError(err).Error("RegisterV2 logTailerInput marshal failed")
			return nil, err
		}
		logTailerConf := monagent.LogTailerConfig{}
		err = yaml.Unmarshal(configData, &logTailerConf)
		if err != nil {
			log.WithError(err).Error("RegisterV2 logTailerInput unmarshal failed")
			return nil, err
		}
		log.Infof("logTailerInput config: %+v", logTailerConf)

		if logTailerConf.RecoveryConfig.Enabled &&
			logTailerConf.RecoveryConfig.LastPositionStoreDir == "" {
			logTailerConf.RecoveryConfig.LastPositionStoreDir = path2.PositionStoreDir()
		}
		logTailer, err := log_tailer.NewLogTailer(logTailerConf)
		if err != nil {
			log.WithError(err).Error("RegisterV2 logTailerInput NewLogTailer failed")
			return nil, err
		}
		return logTailer, nil
	})

}
