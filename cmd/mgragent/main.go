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

package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/api/web"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/mgragent"
	configsdk "github.com/oceanbase/obagent/config/sdk"
	"github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/lib/shellf"
	"github.com/oceanbase/obagent/lib/trace"
)

const (
	ManagerPortKey   = "ocp.agent.manager.http.port"
	OcpAgentHomePath = "obagent.home.path"
)

// command-line arguments
type arguments struct {
	ConfigFile string
}

func run(args arguments) {
	ctx := trace.ContextWithRandomTraceId()

	// Set the umask of the agent process to 0022 so that the operations
	// of the agent are not affected by the umask set by the system
	syscall.Umask(0022)

	conf := mgragent.NewManagerAgentConfig(args.ConfigFile)
	// Replace the home.path in the configuration file with the actual value
	pathMap := map[string]string{OcpAgentHomePath: path.AgentDir()}
	_, err := config.ReplaceConfValues(conf, pathMap)
	if err != nil {
		log.WithError(err).Fatalf("parse mgragent config file path fialed %s", err)
		os.Exit(1)
	}

	// Initialize configuration
	initModuleConfigs(ctx, conf.SDKConfig)

	// Obtain the service startup port based on the port value in the config file
	managerPort := config.GetConfigPropertiesByKey(ManagerPortKey)
	portMap := map[string]string{ManagerPortKey: managerPort}
	_, err = config.ReplaceConfValues(conf, portMap)
	if err != nil {
		log.WithError(err).Fatalf("parse mgragent config file port fialed %s", err)
		os.Exit(1)
	}

	shellf.InitShelf(conf.ShellfConfig.TemplatePath)

	log.WithContext(ctx).Infof("starting ocp manager agent, version %v", config.AgentVersion)
	log.WithContext(ctx).Infof("agent running in %v mode", config.Mode)

	server := web.NewServer(config.Mode, conf.Server)
	go server.Run()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-ch:
		log.WithContext(ctx).Infof("signal '%s' received. exiting...", sig.String())
		server.Stop()
	}
}

func parseArgsAndRun() error {
	var args arguments
	rootCmd := &cobra.Command{
		Use:   "ob_mgragent",
		Short: "OB manager agent",
		Long: "OB manager agent is an operation and maintenance process of OB-Agent," +
			" providing basic host operation and maintenance commands, OB operation and maintenance commands",
	}
	confPath := filepath.Join(path.ConfDir(), "mgragent.yaml")
	rootCmd.PersistentFlags().StringVarP(&args.ConfigFile, "config", "c", confPath, "config file")
	rootCmd.Run = func(cmd *cobra.Command, positionalArgs []string) {
		// The startup phase is set to debug
		log.SetLevel(log.DebugLevel)
		run(args)
	}
	return rootCmd.Execute()
}

func main() {
	runtime.GOMAXPROCS(1)
	err := parseArgsAndRun()
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}

func initModuleConfigs(ctx context.Context, sdkconf config.SDKConfig) {
	// Initialize the SDK
	mgragent.GlobalConfigManager = mgragent.NewManager(mgragent.ManagerConfig{
		ModuleConfigDir:     sdkconf.ModuleConfigDir,
		ConfigPropertiesDir: sdkconf.ConfigPropertiesDir,
	})
	err := configsdk.InitSDK(ctx, sdkconf)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}
	err = configsdk.RegisterMgragentCallbacks(ctx)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}
	err = configsdk.RegisterMonagentCallbacks(ctx)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// Initialize the log
	if err := config.InitModuleConfig(ctx, config.ManagerLogConfigModule); err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// Initialize sock proxy
	if err := config.InitModuleConfig(ctx, config.ManagerAgentProxyConfigModule); err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// Self-monitoring configuration
	if err := config.InitModuleTypeConfig(ctx, config.StatConfigModuleType); err != nil {
		log.WithContext(ctx).Fatalf("init module type %s, err:%+v", config.StatConfigModuleType, err)
	}

	// basic authentication gets the configuration and initializes
	common.InitBasicAuthConf(ctx)

	//Initialize the ob_logcleaner
	if err := config.InitModuleConfig(ctx, config.OBLogcleanerModule); err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// Initialize the logQuerier
	if err := config.InitModuleConfig(ctx, config.ManagerLogQuerierModule); err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// config meta
	err = config.InitModuleConfig(ctx, config.ManagerAgentConfigMetaModule)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}
}
