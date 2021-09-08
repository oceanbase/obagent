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

package config

import "github.com/oceanbase/obagent/lib/crypto"

var (
	CurProcess Process

	AgentVersion string

	Mode AgentMode
)

type AgentMode = string

type Process = string

const (
	ProcessMonitorAgent Process = "monagent"
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

type BasicAuthConfig struct {
	Auth     string `yaml:"auth"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
