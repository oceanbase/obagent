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
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/agentd"
	config2 "github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/lib/path"
	agentLog "github.com/oceanbase/obagent/log"
)

// command-line arguments
type arguments struct {
	ConfigFile string
}

func main() {
	runtime.GOMAXPROCS(1)
	confPath := filepath.Join(path.ConfDir(), "/agentd.yaml")
	rootCmd := &cobra.Command{
		Use:   "ob_agentd",
		Short: "OB agent supervisor",
		Long: "OB agentd is the daemon of ob-Agent. Responsible for starting and stopping," +
			" guarding other agent processes, and status query.",
	}
	rootCmd.PersistentFlags().StringVarP(&confPath, "config", "c", confPath, "config file")
	rootCmd.Run = func(cmd *cobra.Command, positionalArgs []string) {
		run(confPath)
	}
	err := rootCmd.Execute()
	if err != nil {
		log.Error("start agentd failed: ", err)
		os.Exit(-1)
	}
}

func run(confPath string) {
	config := loadConfig(confPath)
	agentLog.InitLogger(agentLog.LoggerConfig{
		Level:      config.LogLevel,
		Filename:   filepath.Join(path.LogDir(), "agentd.log"), //fmt.Sprintf("agentd.%d.log", os.Getpid())),
		MaxSize:    100 * 1024 * 1024,
		MaxBackups: 10,
	})
	log.Infof("starting agentd with config %s", confPath)

	// Replace the home.path in the configuration file with the actual value
	pathMap := map[string]string{"obagent.home.path": path.AgentDir()}
	_, err := config2.ReplaceConfValues(&config, pathMap)
	if err != nil {
		log.Errorf("start agentd with config file '%s' failed: %v", confPath, err)
		os.Exit(-1)
		return
	}

	watchdog := agentd.NewAgentd(config)
	err = watchdog.Start()
	if err != nil {
		log.Errorf("start agentd with config file '%s' failed: %v", confPath, err)
		os.Exit(-1)
		return
	}
	watchdog.ListenSignal()
}

func loadConfig(confPath string) agentd.Config {
	config := agentd.Config{
		LogLevel:        "info",
		LogDir:          "/tmp",
		CleanupDangling: true,
	}

	confFile, err := os.Open(confPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "open config file %s failed: %v\n", confPath, err)
		os.Exit(1)
		return config
	}
	defer confFile.Close()
	err = yaml.NewDecoder(confFile).Decode(&config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "read config file %s failed: %v\n", confPath, err)
		os.Exit(1)
		return config
	}
	return config
}
