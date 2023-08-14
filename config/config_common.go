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

package config

import (
	"strconv"
	"time"

	"github.com/oceanbase/obagent/lib/crypto"
)

var (
	CurProcess     Process
	AgentVersion   string
	Mode           AgentMode = ReleaseMode
	BuildEpoch     string
	BuildGoVersion string
)

var (
	GitBranch        string
	GitCommitId      string
	GitShortCommitId string
	GitCommitTime    string
)

type AgentMode = string

const (
	DebugMode   AgentMode = "debug"
	ReleaseMode AgentMode = "release"
)

type Process = string

const (
	ProcessManagerAgent Process = "ob_mgragent"
	ProcessMonitorAgent Process = "ob_monagent"
	ProcessAgentCtl     Process = "ob_agentctl"
)

type InstallConfig struct {
	Path string `yaml:"path"`
}

type LogConfig struct {
	// log level
	Level string `yaml:"level"`
	// log filename
	Filename string `yaml:"filename"`
	// maxsize of log file
	MaxSize int `yaml:"maxsize"`
	// maxage of log file
	MaxAge int `yaml:"maxage"`
	// max backups of log files
	MaxBackups int  `yaml:"maxbackups"`
	LocalTime  bool `yaml:"localtime"`
	// compress log file
	Compress bool `yaml:"compress"`
}

// Log log wrapper
type Log struct {
	Log LogConfig `yaml:"log"`
}

type SDKConfig struct {
	ConfigPropertiesDir string              `yaml:"configPropertiesDir"`
	ModuleConfigDir     string              `yaml:"moduleConfigDir"`
	CryptoPath          string              `yaml:"cryptoPath"`
	CryptoMethod        crypto.CryptoMethod `yaml:"cryptoMethod"`
}

type ShellfConfig struct {
	// shell template config file path
	TemplatePath string `yaml:"template"`
}

type BasicAuthConfig struct {
	Auth              string `yaml:"auth"`
	MetricAuthEnabled bool   `yaml:"metricAuthEnabled"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
}

func GetAgentInfo() map[string]interface{} {
	var infoMap = map[string]interface{}{
		"name":           CurProcess,
		"version":        AgentVersion,
		"mode":           Mode,
		"buildGoVersion": BuildGoVersion,
	}
	if epoch, err := strconv.Atoi(BuildEpoch); err == nil {
		buildTime := time.Unix(int64(epoch), 0)
		infoMap["buildTime"] = buildTime
	}
	return infoMap
}

func GetGitInfo() map[string]interface{} {
	return map[string]interface{}{
		"branch":        GitBranch,
		"commitId":      GitCommitId,
		"shortCommitId": GitShortCommitId,
		"commitTime":    GitCommitTime,
	}
}
