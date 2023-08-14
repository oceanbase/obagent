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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	json "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/agentd/api"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/agentctl"
	"github.com/oceanbase/obagent/config/sdk"
	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/lib/mask"
	"github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/lib/trace"
	agentlog "github.com/oceanbase/obagent/log"
)

const (
	commandNameInfo    = "info"
	commandNameGitInfo = "git-info"
)

var (
	agentCtlConfig *agentctl.AgentctlConfig
	// root command
	agentCtlCommand = &cobra.Command{
		Use:   "ob_agentctl",
		Short: "ob_agentctl is a CLI for agent management.",
		Long:  `ob_agentctl is a command line tool for the agent. It provides operation, maintenance, and management functions for the agent.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Use == commandNameInfo || cmd.Use == commandNameGitInfo {
				return
			}
			ctx := trace.ContextWithRandomTraceId()
			ctlConfig := cmd.Flag("config").Value.String()
			var err error
			agentCtlConfig, err = LoadConfig(ctlConfig)
			if err != nil {
				setResult(err)
				os.Exit(1)
			}

			// Replace the home.path in the configuration file with the actual value
			pathMap := map[string]string{"obagent.home.path": path.AgentDir()}
			_, err = config.ReplaceConfValues(agentCtlConfig, pathMap)
			if err != nil {
				setResult(err)
				os.Exit(1)
			}

			InitLog(agentCtlConfig.Log)

			err = sdk.InitSDK(ctx, agentCtlConfig.SDKConfig)
			if err != nil {
				setResult(err)
				os.Exit(1)
			}

			err = sdk.RegisterMgragentCallbacks(ctx)
			if err != nil {
				setResult(err)
				os.Exit(1)
			}

			err = sdk.RegisterMonagentCallbacks(ctx)
			if err != nil {
				setResult(err)
				os.Exit(1)
			}
			// Initialize sock proxy
			if err := config.InitModuleConfig(ctx, config.ManagerAgentProxyConfigModule); err != nil {
				setResult(err)
				os.Exit(1)
			}
		},
	}
	setResultOnce sync.Once
)

// Defines the configuration of subcommands and command line parameters
func defineConfigCommands() {
	// config command
	configCommand := &cobra.Command{
		Use:     "config",
		Short:   "config management",
		Long:    "update key-value configs, save configs to config properties files, and notify configs to the module using these configs",
		Example: "config --update key1=value1,key2=value2",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := trace.ContextWithRandomTraceId()
			// config meta
			if err := config.InitModuleTypeConfig(ctx, config.ConfigMetaModuleType); err != nil {
				log.WithContext(ctx).Fatal(err)
			}

			updateConfigs, err := cmd.Flags().GetStringSlice("update")
			if err != nil {
				setResult(err)
				log.WithContext(ctx).WithField("args", os.Args).Fatalf("agentctl config --update err:%s", err)
			}
			notifyModules, err := cmd.Flags().GetStringSlice("notify")
			if err != nil {
				setResult(err)
				log.WithContext(ctx).WithField("args", os.Args).Fatalf("agentctl config --notify err:%s", err)
			}
			validateConfigs, err := cmd.Flags().GetStringSlice("validate")
			if err != nil {
				setResult(err)
				log.WithContext(ctx).WithField("args", os.Args).Fatalf("agentctl config --validate err:%s", err)
			}

			log.WithContext(ctx).Infof("agentctl config updates:%+v, notify modules:%+v, validate configs:%+v", mask.MaskSlice(updateConfigs), notifyModules, mask.MaskSlice(validateConfigs))

			err = runUpdateConfigs(ctx, updateConfigs)
			if err != nil {
				log.WithContext(ctx).WithField("updateConfigs", mask.MaskSlice(updateConfigs)).Errorf("agentctl config update config err:%s", err)
				setResult(err)
			}

			err = runNotifyModules(ctx, notifyModules)
			if err != nil {
				log.WithContext(ctx).WithField("notifyModules", notifyModules).Errorf("agentctl config notify modules err:%s", err)
				setResult(err)
			}

			err = runValidateConfigs(ctx, validateConfigs)
			if err != nil {
				log.WithContext(ctx).WithField("runValidateConfigs", mask.MaskSlice(validateConfigs)).Errorf("agentctl config validate err:%s", err)
			}
			setResult(err)
		},
	}
	// update configuration: The key-value pair is used for update.
	// You need to enter the complete configuration.
	// The UPDATE verifies the configuration, saves the configuration,
	// and notifies services to use the configuration.
	configCommand.PersistentFlags().StringSliceP("update", "u", nil, "key-value pairs, e.g., key1=value1,key2=value2")
	// Notification service configuration takes effect.
	configCommand.PersistentFlags().StringSliceP("notify", "n", nil, "notify modules, e.g., mgragent.config,monagent")
	// Verify that the configurations are consistent
	configCommand.PersistentFlags().StringSliceP("validate", "v", nil, "validate config key-value pairs, e.g., key1=value1,key2=value2")

	configNotifyCommand := &cobra.Command{
		Use:     "notify",
		Short:   "notify config change",
		Long:    "notify modules configs changes. omitting modules means notify all modules",
		Example: "notify module1 module2",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := trace.ContextWithRandomTraceId()
			var err error
			if len(args) > 0 {
				err = config.NotifyModules(ctx, args)
			} else {
				err = config.NotifyAllModules(ctx)
			}
			if err != nil {
				log.WithContext(ctx).WithField("args", os.Args).Fatalf("agentctl config notify err:%s", err)
			}
			setResult(err)
		},
	}
	configChangeCommand := &cobra.Command{
		Use:     "change",
		Short:   "change config properties",
		Long:    "change config properties",
		Example: "change k1=v1 k2=v2",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := trace.ContextWithRandomTraceId()
			err := runUpdateConfigs(ctx, args)
			if err != nil {
				log.WithContext(ctx).WithField("args", os.Args).Fatalf("agentctl config update err:%s", err)
			}
			setResult(err)
		},
	}
	validateChangeCommand := &cobra.Command{
		Use:     "validate",
		Short:   "validate config properties",
		Long:    "validate config properties",
		Example: "validate k1=v1 k2=v2",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := trace.ContextWithRandomTraceId()
			err := runValidateConfigs(ctx, args)
			if err != nil {
				log.WithContext(ctx).WithField("args", os.Args).Fatalf("agentctl config validate err:%s", err)
			}
			setResult(err)
		},
	}
	configCommand.AddCommand(configChangeCommand, configNotifyCommand, validateChangeCommand)

	agentCtlCommand.AddCommand(configCommand)
}

func adminConf() agent.AdminConf {
	return agent.AdminConf{
		RunDir:           agentCtlConfig.RunDir,
		LogDir:           agentCtlConfig.LogDir,
		ConfDir:          agentCtlConfig.ConfDir,
		BackupDir:        agentCtlConfig.BackupDir,
		TempDir:          agentCtlConfig.TempDir,
		PkgStoreDir:      agentCtlConfig.PkgStoreDir,
		TaskStoreDir:     agentCtlConfig.TaskStoreDir,
		AgentPkgName:     agentCtlConfig.AgentPkgName,
		PkgExt:           agentCtlConfig.PkgExt,
		StartWaitSeconds: 10,
		StopWaitSeconds:  10,
		AgentdPath:       path.AgentdPath(),
	}
}

func defineInfoCommands() {
	agentCtlCommand.AddCommand(&cobra.Command{
		Use: commandNameInfo,
		Run: func(cmd *cobra.Command, args []string) {
			onSuccess(config.GetAgentInfo())
		},
	})
	agentCtlCommand.AddCommand(&cobra.Command{
		Use: commandNameGitInfo,
		Run: func(cmd *cobra.Command, args []string) {
			onSuccess(config.GetGitInfo())
		},
	})
}

func defineOperationCommands() {
	agentCtlCommand.AddCommand(&cobra.Command{
		Use: "status",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				onError(errors.New("too many arguments"))
				return
			}
			admin := agent.NewAdmin(adminConf())
			status, err := admin.AgentStatus()
			if err != nil {
				onError(err)
			} else {
				onSuccess(status)
			}
		},
	})
	agentCtlCommand.AddCommand(&cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				onError(errors.New("too many arguments"))
				return
			}
			admin := agent.NewAdmin(adminConf())
			err := admin.StartAgent()
			if err != nil {
				onError(err)
			} else {
				onSuccess("ok")
			}
		},
	})
	stopCommand := &cobra.Command{
		Use: "stop",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				onError(errors.New("too many arguments"))
				return
			}
			taskToken := cmd.Flag("task-token").Value.String()
			admin := agent.NewAdmin(adminConf())
			err := admin.StopAgent(agent.TaskToken{TaskToken: taskToken})
			if err != nil {
				onError(err)
			} else {
				onSuccess("ok")
			}
		},
	}
	stopCommand.PersistentFlags().String("task-token", "", "task token to store result")
	agentCtlCommand.AddCommand(stopCommand)

	serviceCommand := &cobra.Command{
		Use: "service",
	}
	serviceStartCommand := &cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				onError(errors.New("missing service name"))
				return
			}
			name := args[0]
			taskToken := cmd.Flag("task-token").Value.String()
			admin := agent.NewAdmin(adminConf())
			err := admin.StartService(agent.StartStopServiceParam{
				TaskToken: agent.TaskToken{
					TaskToken: taskToken,
				},
				StartStopAgentParam: api.StartStopAgentParam{
					Service: name,
				},
			})
			if err != nil {
				onError(err)
			} else {
				onSuccess("ok")
			}
		},
	}
	serviceStartCommand.PersistentFlags().String("task-token", "", "task token to store result")

	serviceStopCommand := &cobra.Command{
		Use: "stop",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				onError(errors.New("missing service name"))
				return
			}
			name := args[0]
			taskToken := cmd.Flag("task-token").Value.String()
			admin := agent.NewAdmin(adminConf())
			err := admin.StopService(agent.StartStopServiceParam{
				TaskToken: agent.TaskToken{
					TaskToken: taskToken,
				},
				StartStopAgentParam: api.StartStopAgentParam{
					Service: name,
				},
			})
			if err != nil {
				onError(err)
			} else {
				onSuccess("ok")
			}
		},
	}
	serviceStopCommand.PersistentFlags().String("task-token", "", "task token to store result")

	serviceCommand.AddCommand(serviceStartCommand)
	serviceCommand.AddCommand(serviceStopCommand)
	agentCtlCommand.AddCommand(serviceCommand)

	restartCommand := &cobra.Command{
		Use: "restart",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				onError(errors.New("too many arguments"))
				return
			}
			taskToken := cmd.Flag("task-token").Value.String()
			admin := agent.NewAdmin(adminConf())
			err := admin.RestartAgent(agent.TaskToken{TaskToken: taskToken})
			if err != nil {
				onError(err)
			} else {
				onSuccess("ok")
			}
		},
	}
	restartCommand.PersistentFlags().String("task-token", "", "task token to store result")
	agentCtlCommand.AddCommand(restartCommand)

	reinstallCommand := &cobra.Command{
		Use: "reinstall",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				onError(errors.New("too many arguments"))
				return
			}
			taskToken := cmd.Flag("task-token").Value.String()
			admin := agent.NewAdmin(adminConf())
			source := cmd.Flag("source").Value.String()
			checksum := cmd.Flag("checksum").Value.String()
			version := cmd.Flag("version").Value.String()
			// todo validate args
			err := admin.ReinstallAgent(agent.ReinstallParam{
				TaskToken: agent.TaskToken{TaskToken: taskToken},
				DownloadParam: agent.DownloadParam{
					Source:   source,
					Checksum: checksum,
					Version:  version,
				},
			})
			if err != nil {
				onError(err)
			} else {
				onSuccess("ok")
			}
		},
	}
	reinstallCommand.PersistentFlags().String("source", "", "package source")
	reinstallCommand.PersistentFlags().String("checksum", "", "package checksum")
	reinstallCommand.PersistentFlags().String("version", "", "package version")
	reinstallCommand.PersistentFlags().String("task-token", "", "task token to store result")
	agentCtlCommand.AddCommand(reinstallCommand)

}

func defineVersionCommands() {
	agentCtlCommand.AddCommand(&cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			onSuccess(config.AgentVersion)
		},
	})
}

func main() {
	runtime.GOMAXPROCS(1)
	confPath := filepath.Join(path.ConfDir(), "agentctl.yaml")
	agentCtlCommand.PersistentFlags().StringP("config", "c", confPath, "config file")
	defineInfoCommands()
	defineOperationCommands()
	defineConfigCommands()
	defineVersionCommands()

	if err := agentCtlCommand.Execute(); err != nil {
		log.WithField("args", os.Args).Fatal(err)
		setResult(err)
	}
}

func LoadConfig(configFile string) (*agentctl.AgentctlConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	ret := &agentctl.AgentctlConfig{
		ConfDir:      path.ConfDir(),
		RunDir:       path.RunDir(),
		LogDir:       path.LogDir(),
		BackupDir:    path.BackupDir(),
		TempDir:      path.TempDir(),
		TaskStoreDir: path.TaskStoreDir(),
		AgentPkgName: filepath.Join(path.AgentDir(), "obagent"),
		PkgExt:       "rpm",
		PkgStoreDir:  path.PkgStoreDir(),
	}
	err = yaml.NewDecoder(f).Decode(ret)
	return ret, err
}

// init log
func InitLog(conf config.LogConfig) {
	agentlog.InitLogger(agentlog.LoggerConfig{
		Level:      conf.Level,
		Filename:   conf.Filename,
		MaxSize:    conf.MaxSize,
		MaxAge:     conf.MaxAge,
		MaxBackups: conf.MaxBackups,
		LocalTime:  conf.LocalTime,
		Compress:   conf.Compress,
	})
}

// set the result to stdout
// logs will be written to log file
func setResult(err error) {
	setResultOnce.Do(func() {
		if err != nil {
			onError(err)
			return
		}
		onSuccess("success")
	})
}

// the returned err will be written to stderr
func onError(err error) {
	resp := &agent.AgentctlResponse{
		Successful: false,
		Error:      err.Error(),
	}
	data, jsonerr := json.Marshal(resp)
	if jsonerr != nil {
		log.WithField("error", err).Errorf("json marshal err:%s", jsonerr)
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(-1)
		return
	}
	log.WithField("response", string(data)).Info("agentctl error")
	fmt.Fprintf(os.Stderr, "%s", data)
	os.Exit(-1)
}

// success message will be written to stdout
func onSuccess(message interface{}) {
	resp := &agent.AgentctlResponse{
		Successful: true,
		Message:    message,
	}
	data, jsonerr := json.Marshal(resp)
	if jsonerr != nil {
		log.WithField("message", message).Errorf("json marshal err:%s", jsonerr)
		return
	}
	log.WithField("response", string(data)).Info("agentctl success")
	fmt.Fprintf(os.Stdout, "%s", data)
}

func runAgentctl(configfile string) error {
	return nil
}

// run validate commands: validate config is identical
func runNotifyModules(ctx context.Context, modules []string) error {
	if len(modules) <= 0 {
		return nil
	}
	return config.NotifyModules(ctx, modules)
}

// run config command: update module config
func runUpdateConfigs(ctx context.Context, pairs []string) error {
	if len(pairs) <= 0 {
		return nil
	}
	return config.UpdateConfigPairs(ctx, pairs)
}

func runValidateConfigs(ctx context.Context, pairs []string) error {
	if len(pairs) <= 0 {
		return nil
	}
	return config.ValidateConfigPairs(ctx, pairs)
}
