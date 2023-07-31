package agentd

import (
	"time"

	"github.com/alecthomas/units"
)

// ServiceConfig sub service config
type ServiceConfig struct {
	//Program can be absolute or relative path or just command name
	Program string `yaml:"program"`
	//Args program arguments
	Args []string `yaml:"args"`
	//Cwd current work dir of service
	Cwd string `yaml:"cwd"`
	//RunDir dir to put something like pid, socket, lock
	RunDir string `yaml:"runDir"`
	//Path to redirect Stdout
	Stdout string `yaml:"stdout"`
	//Path to redirect Stderr
	Stderr string `yaml:"stderr"`
	//When service quited too quickly for QuickExitLimit times, service will not restart even more.
	QuickExitLimit int `yaml:"quickExitLimit"`
	//Services lives less then MinLiveTime will treated as quited too quickly.
	MinLiveTime time.Duration `yaml:"minLiveTime"`
	//KillWait after stop service and process not exited, will send SIGKILL signal to it.
	//0 means no wait and won't send SIGKILL signal.
	KillWait time.Duration `yaml:"killWait"`
	//FinalWait after stop service and process not exited, will not wait for it. and return an error
	FinalWait time.Duration `yaml:"finalWait"`
	//Limit cpu and memory usage
	Limit LimitConfig `yaml:"limit"`
}

// Config agentd config
type Config struct {
	//RunDir dir to put something like pid, socket, lock
	RunDir string `yaml:"runDir"`
	//LogDir dir to write agentd log
	LogDir string `yaml:"logDir"`
	//LogLevel highest level to log
	LogLevel string `yaml:"logLevel"`
	//Services sub services config
	Services map[string]ServiceConfig `yaml:"services"`
	//CleanupDangling whether cleanup dangling service process or not
	CleanupDangling bool
}

type LimitConfig struct {
	//CpuQuota max cpu usage percentage. 1.0 means 100%, 2.0 means 200%
	CpuQuota float32 `yaml:"cpuQuota"`
	//MemoryQuota max memory limit in bytes
	MemoryQuota units.Base2Bytes `yaml:"memoryQuota"`
}
