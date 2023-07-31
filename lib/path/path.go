package path

import (
	"os"
	"path/filepath"
)

const (
	Agentd   = "ob_agentd"
	MgrAgent = "ob_mgragent"
	MonAgent = "ob_monagent"
	Agentctl = "ob_agentctl"
)

func MyPath() string {
	ret, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	return ret
}

func ProgramName() string {
	return filepath.Base(os.Args[0])
}

func BinDir() string {
	return filepath.Dir(MyPath())
}

func AgentDir() string {
	return filepath.Dir(BinDir())
}

func RunDir() string {
	return filepath.Join(AgentDir(), "run")
}

func ConfDir() string {
	return filepath.Join(AgentDir(), "conf")
}

func TempDir() string {
	return filepath.Join(AgentDir(), "tmp")
}

func LogDir() string {
	return filepath.Join(AgentDir(), "log")
}

func BackupDir() string {
	return filepath.Join(AgentDir(), "backup")
}

func PkgStoreDir() string {
	return filepath.Join(AgentDir(), "pkg_store")
}

func TaskStoreDir() string {
	return filepath.Join(AgentDir(), "task_store")
}

func PositionStoreDir() string {
	return filepath.Join(AgentDir(), "position_store")
}

func AgentdPath() string {
	return filepath.Join(BinDir(), Agentd)
}

func MgrAgentPath() string {
	return filepath.Join(BinDir(), MgrAgent)
}

func MonAgentPath() string {
	return filepath.Join(BinDir(), MonAgent)
}

func AgentctlPath() string {
	return filepath.Join(BinDir(), Agentctl)
}
