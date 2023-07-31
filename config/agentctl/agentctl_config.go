package agentctl

import "github.com/oceanbase/obagent/config"

// agentctl meta
type AgentctlConfig struct {
	SDKConfig config.SDKConfig `yaml:"sdkConfig"`
	// log config
	Log config.LogConfig `yaml:"log"`

	RunDir       string `yaml:"runDir"`
	ConfDir      string `yaml:"confDir"`
	LogDir       string `yaml:"logDir"`
	BackupDir    string `yaml:"backupDir"`
	TempDir      string `yaml:"tempDir"`
	TaskStoreDir string `yaml:"taskStoreDir"`
	AgentPkgName string `yaml:"agentPkgName"`
	PkgExt       string `yaml:"pkgExt"`
	PkgStoreDir  string `yaml:"pkgStoreDir"`
}
