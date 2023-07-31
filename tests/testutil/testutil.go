package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

var (
	RunDir              string
	TmpRootDir          string
	ConfDir             string
	ConfigPropertiesDir string
	ModuleConfigDir     string
	BinDir              string
	LogDir              string
	TempDir             string
	BackupDir           string
	PkgStoreDir         string
	TaskStoreDir        string

	ProjectRoot   string
	MockAgentPath string
	AgentdPath    string
	MgrAgentPath  string
	MonAgentPath  string
	AgentctlPath  string
)

func init() {
	_, testFile, _, _ := runtime.Caller(0)
	fmt.Println(testFile)
	ProjectRoot = filepath.Dir(filepath.Dir(filepath.Dir(testFile)))
	BinDir = filepath.Join(ProjectRoot, "bin")
	TmpRootDir = filepath.Join("/tmp", fmt.Sprintf("ocp_test_%d", time.Now().UnixNano()))
	ConfDir = filepath.Join(TmpRootDir, "conf")
	ConfigPropertiesDir = filepath.Join(ConfDir, "config_properties")
	ModuleConfigDir = filepath.Join(ConfDir, "module_config")
	TempDir = filepath.Join(TmpRootDir, "tmp")
	RunDir = filepath.Join(TmpRootDir, "run")
	LogDir = filepath.Join(TmpRootDir, "log")
	BackupDir = filepath.Join(TmpRootDir, "backup")
	PkgStoreDir = filepath.Join(TmpRootDir, "pkg_store")
	TaskStoreDir = filepath.Join(TmpRootDir, "task_store")

	MockAgentPath = filepath.Join(BinDir, "mock_agent")
	AgentdPath = filepath.Join(BinDir, "ob_agentd")
	MgrAgentPath = filepath.Join(BinDir, "ob_monagent")
	MonAgentPath = filepath.Join(BinDir, "ob_mgragent")
	AgentctlPath = filepath.Join(BinDir, "ob_agentctl")
}

func MakeDirs() {
	dirsToMake := []string{
		RunDir, ConfDir, ConfigPropertiesDir, ModuleConfigDir, LogDir, BackupDir, TempDir, PkgStoreDir, TaskStoreDir,
	}
	for _, dir := range dirsToMake {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("make test dir '%s' failed %v", dir, err)
		}
	}
}

func DelTestFiles() {
	err := os.RemoveAll(TmpRootDir)
	if err != nil {
		fmt.Printf("del all test dir '%s' failed %v", TmpRootDir, err)
	}
}

func KillAll() {
	ps, err := process.Processes()
	if err != nil {
		return
	}
	for _, p := range ps {
		name, err := p.Name()
		if err != nil {
			continue
		}
		if name == "ob_agentd" || name == "ob_mgragent" || name == "ob_monagent" || name == "ob_agentctl" || name == "mock_agent" {
			exe, err := p.Exe()
			if err != nil {
				continue
			}
			if !strings.HasPrefix(exe, ProjectRoot) {
				continue
			}
			fmt.Printf("kill process %s %d\n", name, p.Pid)
			_ = p.Kill()
		}
	}
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func BuildBins() error {
	if FileExists(AgentdPath) && FileExists(MonAgentPath) && FileExists(MgrAgentPath) && FileExists(AgentctlPath) {
		return nil
	}
	fmt.Printf("build agent in '%s'\n", ProjectRoot)
	_ = os.Chdir(ProjectRoot)
	out, err := exec.Command("make", "pre-build", "build").CombinedOutput()
	if err != nil {
		fmt.Printf("build agent binaries failed %v\n%s\n", err, string(out))
		return err
	}
	return nil
}

func BuildMockAgent() error {
	_, testFile, _, _ := runtime.Caller(0)
	mockAgentGo := filepath.Join(filepath.Dir(filepath.Dir(testFile)), "mockagent.go")
	if !FileExists(MockAgentPath) {
		out, err := exec.Command("go", "build", "-o", MockAgentPath, mockAgentGo).CombinedOutput()
		if err != nil {
			fmt.Printf("build mock agent failed %v\n%s\n", err, string(out))
			return err
		}
	}
	return nil
}
