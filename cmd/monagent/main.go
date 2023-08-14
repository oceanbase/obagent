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

package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/oceanbase/obagent/api/web"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/config/sdk"
	"github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/lib/trace"
	"github.com/oceanbase/obagent/monitor/engine"
	_ "github.com/oceanbase/obagent/monitor/plugins/exporters"
	_ "github.com/oceanbase/obagent/monitor/plugins/inputs"
	_ "github.com/oceanbase/obagent/monitor/plugins/outputs"
	_ "github.com/oceanbase/obagent/monitor/plugins/processors"
)

const (
	MonitorPortKey   = "ocp.agent.monitor.http.port"
	OcpAgentHomePath = "obagent.home.path"
)

var (
	// root command
	monagentCommand = &cobra.Command{
		Use:   "ob_monagent",
		Short: "ob_monagent is a monitoring agent.",
		Long:  `ob_monagent is a monitoring agent for gathering, processing and pushing monitor metrics.`,
		Run: func(cmd *cobra.Command, args []string) {
			// The startup phase is set to debug
			log.SetLevel(log.DebugLevel)
			err := runMonitorAgent()
			if err != nil {
				log.WithField("args", args).Errorf("monagent execute err:%s", err)
			}
		},
	}
)

func init() {
	confPath := filepath.Join(path.ConfDir(), "monagent.yaml")
	// monagent server config file
	monagentCommand.PersistentFlags().StringP("config", "c", confPath, "config file")
	// plugins config use dir, all yaml files in the dir will be used as plugin config file
	monagentCommand.PersistentFlags().StringP("pipelines_config_dir", "d", "conf/monagent/pipelines", "monitor pipelines config file dir")

	_ = viper.BindPFlag("config", monagentCommand.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("pipelines_config_dir", monagentCommand.PersistentFlags().Lookup("pipelines_config_dir"))
}

func main() {
	runtime.GOMAXPROCS(1)
	if err := monagentCommand.Execute(); err != nil {
		log.WithField("args", os.Args).Errorf("monagentCommand Execute failed %s", err.Error())
		os.Exit(-1)
	}
}

func runMonitorAgent() error {
	monagentConfig, err := monagent.DecodeMonitorAgentServerConfig(viper.GetString("config"))
	if err != nil {
		return errors.Wrap(err, "read monitor agent server config")
	}

	// Replace the home.path in the configuration file with the actual value
	contextMap := map[string]string{OcpAgentHomePath: path.AgentDir()}
	_, err = config.ReplaceConfValues(monagentConfig, contextMap)
	if err != nil {
		return errors.Wrap(err, "Failed to parse config file path")
	}

	ctxlog := trace.ContextWithRandomTraceId()
	err = sdk.InitSDK(ctxlog, config.SDKConfig{
		ModuleConfigDir:     monagentConfig.ModulePath,
		ConfigPropertiesDir: monagentConfig.PropertiesPath,
		CryptoPath:          monagentConfig.CryptoPath,
		CryptoMethod:        monagentConfig.CryptoMethod,
	})
	log.WithContext(ctxlog).Infof("sdk inited")
	if err != nil {
		return errors.Wrap(err, "init config sdk")
	}

	// Obtain the service startup port based on the port value in the config file
	monitorPort := config.GetConfigPropertiesByKey(MonitorPortKey)
	portMap := map[string]string{MonitorPortKey: monitorPort}
	_, err = config.ReplaceConfValues(monagentConfig, portMap)
	if err != nil {
		return errors.Wrap(err, "Failed to parse config file port")
	}

	monagentServer := web.NewMonitorAgentServer(monagentConfig)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(time.Second * 3)
	}()

	// start manager
	engine.GetPipelineManager().Schedule(ctx)
	engine.GetConfigManager().Schedule(ctx)

	err = sdk.RegisterMonagentCallbacks(ctxlog)
	if err != nil {
		return errors.Wrap(err, "register monagent callbacks")
	}

	// Initialize the log
	if err := config.InitModuleTypeConfig(ctxlog, config.MonitorLogConfigModuleType); err != nil {
		log.WithContext(ctxlog).WithError(err).Errorf("init module type %s, err:%+v", config.MonitorLogConfigModuleType, err)
	}
	// Initialize self-monitoring
	if err := config.InitModuleTypeConfig(ctxlog, config.StatConfigModuleType); err != nil {
		log.WithContext(ctxlog).WithError(err).Errorf("init module type %s, err:%+v", config.StatConfigModuleType, err)
	}
	// Initialize basic authentication
	if err := config.InitModuleTypeConfig(ctxlog, config.MonitorServerBasicAuthModuleType); err != nil {
		log.WithContext(ctxlog).WithError(err).Errorf("init module type %s, err:%+v", config.MonitorServerBasicAuthModuleType, err)
	}
	// Initialize monitoring and collection pipeline configuration
	if err := config.InitModuleTypeConfig(ctxlog, config.MonitorPipelineModuleType); err != nil {
		log.WithContext(ctxlog).WithError(err).Errorf("init module type %s, err:%+v", config.MonitorPipelineModuleType, err)
	}
	// Initialize meta config
	if err := config.InitModuleConfig(ctx, config.MonitorAgentConfigMetaModule); err != nil {
		log.WithContext(ctxlog).WithError(err).Errorf("init module type %s, err:%+v", config.ConfigMetaModuleType, err)
	}

	go func() {
		monagentServer.RegisterRouter()
		monagentServer.Run()
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-ch:
		log.WithContext(ctx).Infof("signal '%s' received. exiting...", sig.String())
		engine.GetPipelineManager().Stop(ctx)
		monagentServer.Server.Cancel()
		close(engine.PipelineRouteChan)
	}

	return nil
}
