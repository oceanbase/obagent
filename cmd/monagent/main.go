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

package main

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/sdk"
	"github.com/oceanbase/obagent/engine"
	agentlog "github.com/oceanbase/obagent/log"
	_ "github.com/oceanbase/obagent/plugins/exporters"
	_ "github.com/oceanbase/obagent/plugins/inputs"
	_ "github.com/oceanbase/obagent/plugins/outputs"
	_ "github.com/oceanbase/obagent/plugins/processors"
)

var (
	// root command
	monagentCommand = &cobra.Command{
		Use:   "monagent",
		Short: "monagent is a monitoring agent.",
		Long:  `monagent is a monitoring agent for gathering, processing and pushing monitor metrics.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := runMonitorAgent()
			if err != nil {
				log.WithField("args", args).Errorf("monagent execute err:%s", err)
			}
		},
	}
)

func init() {
	// monagent server config file
	monagentCommand.PersistentFlags().StringP("config", "c", "conf/monagent.yaml", "config file")
	// plugins config use dir, all yaml files in the dir will be used as plugin config file
	monagentCommand.PersistentFlags().StringP("pipelines_config_dir", "d", "conf/monagent/pipelines", "monitor pipelines config file dir")

	_ = viper.BindPFlag("config", monagentCommand.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("pipelines_config_dir", monagentCommand.PersistentFlags().Lookup("pipelines_config_dir"))
}

func main() {
	if err := monagentCommand.Execute(); err != nil {
		log.WithField("args", os.Args).Errorf("monagentCommand Execute failed %s", err.Error())
	}
}

func runMonitorAgent() error {
	monagentConfig, err := config.DecodeMonitorAgentServerConfig(viper.GetString("config"))
	if err != nil {
		return errors.Wrap(err, "read monitor agent server config")
	}

	// init log for monagent
	agentlog.InitLogger(agentlog.LoggerConfig{
		Level:      monagentConfig.Log.Level,
		Filename:   monagentConfig.Log.Filename,
		MaxSize:    monagentConfig.Log.MaxSize,
		MaxAge:     monagentConfig.Log.MaxAge,
		MaxBackups: monagentConfig.Log.MaxBackups,
		LocalTime:  monagentConfig.Log.LocalTime,
		Compress:   monagentConfig.Log.Compress,
	})

	monagentServer := engine.NewMonitorAgentServer(monagentConfig)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(time.Second * 3)
	}()

	// start manager
	engine.GetPipelineManager().Schedule(ctx)
	engine.GetConfigManager().Schedule(ctx)

	err = sdk.InitSDK(config.SDKConfig{
		ModuleConfigDir:     monagentConfig.ModulePath,
		ConfigPropertiesDir: monagentConfig.PropertiesPath,
		CryptoPath:          monagentConfig.CryptoPath,
		CryptoMethod:        monagentConfig.CryptoMethod,
	})
	log.Infof("sdk inited")
	if err != nil {
		return errors.Wrap(err, "init config sdk")
	}
	err = sdk.RegisterMonagentCallbacks()
	if err != nil {
		return errors.Wrap(err, "register monagent callbacks")
	}

	err = config.InitModuleTypeConfig(ctx, config.MonitorAdminBasicAuthModuleType)
	err = config.InitModuleTypeConfig(ctx, config.MonitorServerBasicAuthModuleType)
	err = config.InitModuleTypeConfig(ctx, config.MonitorPipelineModuleType)
	if err != nil {
		log.WithError(err).Errorf("init pipeline config, err:%+v", err)
	}

	err = monagentServer.RegisterRouter()
	if err != nil {
		return errors.Wrap(err, "monitor agent server register route")
	}

	err = monagentServer.Run()
	if err != nil {
		return errors.Wrap(err, "start monitor agent server")
	}

	return nil
}
