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

package agent

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shirou/gopsutil/v3/process"

	"github.com/oceanbase/obagent/lib/command"
	"github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/tests/testutil"
)

func writeConfig() {
	agentdConfTpl := `
runDir: %RUN_DIR%
logDir: %LOG_DIR%
services:
  ob_mgragent:
    program: %MOCK_AGENT%
    args: [ ob_mgragent, %RUN_DIR%, 1, 1 ]
    runDir: %RUN_DIR%
    #killWait: 0s
    finalWait: 2s
    minLiveTime: 5s
    quickExitLimit: 3
    stdout: %LOG_DIR%/mockmgr.output.log
    stderr: %LOG_DIR%/mockmgr.error.log

  ob_monagent:
    program: %MOCK_AGENT%
    args: [ ob_monagent, %RUN_DIR%, 1, 1 ]
    runDir: %RUN_DIR%
    #killWait: 0s
    finalWait: 2s
    minLiveTime: 5s
    quickExitLimit: 3
    stdout: %LOG_DIR%/mockmon.output.log
    stderr: %LOG_DIR%/mockmon.error.log
`
	r := strings.NewReplacer(
		"%RUN_DIR%", testutil.RunDir,
		"%LOG_DIR%", testutil.LogDir,
		"%MOCK_AGENT%", testutil.MockAgentPath,
	)
	agentdConfContent := r.Replace(agentdConfTpl)
	fmt.Println(agentdConfContent)
	_ = ioutil.WriteFile(filepath.Join(testutil.ConfDir, "agentd.yaml"), []byte(agentdConfContent), 0644)
}

func writeFailConfig() {
	agentdConfTpl := `
runDir: %RUN_DIR%
logDir: %LOG_DIR%
services:
  ob_mgragent:
    program: %MOCK_AGENT%
    args: [ ob_mgragent, %RUN_DIR%, -1, 1 ]
    runDir: %RUN_DIR%
    #killWait: 0s
    finalWait: 2s
    minLiveTime: 5s
    quickExitLimit: 3
    stdout: %LOG_DIR%/mockmgr.output.log
    stderr: %LOG_DIR%/mockmgr.error.log

  ob_monagent:
    program: %MOCK_AGENT%
    args: [ ob_monagent, %RUN_DIR%, 1, 1 ]
    runDir: %RUN_DIR%
    #killWait: 0s
    finalWait: 2s
    minLiveTime: 5s
    quickExitLimit: 3
    stdout: %LOG_DIR%/mockmon.output.log
    stderr: %LOG_DIR%/mockmon.error.log
`
	r := strings.NewReplacer(
		"%RUN_DIR%", testutil.RunDir,
		"%LOG_DIR%", testutil.LogDir,
		"%MOCK_AGENT%", testutil.MockAgentPath,
	)
	agentdConfContent := r.Replace(agentdConfTpl)
	fmt.Println(agentdConfContent)
	_ = ioutil.WriteFile(filepath.Join(testutil.ConfDir, "agentd.yaml"), []byte(agentdConfContent), 0644)
}

func TestMain(m *testing.M) {
	err := testutil.BuildBins()
	if err != nil {
		os.Exit(2)
	}
	err = testutil.BuildMockAgent()
	if err != nil {
		os.Exit(2)
	}

	ret := m.Run()

	testutil.KillAll()
	testutil.DelTestFiles()
	os.Exit(ret)
}

var adminConf = AdminConf{
	RunDir:           testutil.RunDir,
	ConfDir:          testutil.ConfDir,
	LogDir:           testutil.LogDir,
	BackupDir:        testutil.BackupDir,
	TempDir:          testutil.TempDir,
	TaskStoreDir:     testutil.TaskStoreDir,
	AgentPkgName:     "obagent",
	PkgStoreDir:      testutil.PkgStoreDir,
	PkgExt:           "rpm",
	StartWaitSeconds: 3,
	StopWaitSeconds:  3,

	AgentdPath: testutil.AgentdPath,
}

func TestAdmin_StartStopAgent(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()
	writeConfig()

	admin := NewAdmin(adminConf)
	err := admin.StartAgent()
	if err != nil {
		t.Error(err)
		return
	}
	err = admin.StopAgent(TaskToken{TaskToken: command.GenerateTaskId()})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestAdmin_StartAgentFail(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()
	writeFailConfig()
	admin := NewAdmin(adminConf)
	err := admin.StartAgent()
	if err == nil {
		t.Error("should timeout")
		return
	}
	fmt.Println(err)
}

func TestAdmin_RestartAgent(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()
	writeConfig()

	admin := NewAdmin(adminConf)
	err := admin.StartAgent()
	if err != nil {
		t.Error(err)
		return
	}

	err = admin.RestartAgent(TaskToken{TaskToken: command.GenerateTaskId()})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestAdmin_RestartAgentFail(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()
	writeConfig()
	admin := NewAdmin(adminConf)
	err := admin.StartAgent()
	if err != nil {
		t.Error(err)
		return
	}
	writeFailConfig()
	err = admin.RestartAgent(TaskToken{TaskToken: "test"})
	if err == nil {
		t.Error("should timeout")
		return
	}
	fmt.Println(err)

	ps, _ := process.Processes()
	n := 0
	for _, p := range ps {
		name, _ := p.Name()
		exe, _ := p.Exe()
		if name == path.Agentd && strings.HasPrefix(exe, testutil.ProjectRoot) {
			n++
		}
	}
	if n != 1 {
		t.Errorf("should only has one live agentd, but %d", n)
	}
}

func TestAdmin_BackupRestoreConfig(t *testing.T) {
	testutil.MakeDirs()
	writeConfig()

	admin := NewAdmin(adminConf)

	opCtx := OpCtx{}

	err := admin.backupConfig(&opCtx)
	if err != nil {
		t.Error(err)
	}
	if opCtx.confBackupDir == "" {
		t.Error("confBackupDir empty")
	}
	fmt.Println(opCtx.confBackupDir)
	if _, err = os.Stat(filepath.Join(opCtx.confBackupDir, "agentd.yaml")); err != nil {
		t.Error("backup not copied files")
	}
	_ = os.MkdirAll(admin.conf.ConfDir, 0755)
	err = admin.updateConfig(&opCtx)
	if err != nil {
		t.Error(err)
	}
	err = admin.restoreConfig(&opCtx)
	if err != nil {
		t.Error(err)
	}
}

func TestAdmin_BackupRestorePid(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()
	writeConfig()

	admin := NewAdmin(adminConf)
	err := admin.StartAgent()
	if err != nil {
		t.Error(err)
		return
	}
	opCtx := OpCtx{}
	err = admin.backupPid(&opCtx)
	if err != nil {
		t.Error(err)
		return
	}
	admin.restorePid(&opCtx)
	_, err = admin.readPid(path.Agentd)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestAdmin_ReinstallAgent(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()
	writeConfig()

	admin := NewAdmin(adminConf)
	admin.installPkgFunc = func(opCtx *OpCtx, pkgPath string) error {
		//admin.restoreConfig(opCtx)
		return nil
	}
	admin.checkCurrentPkgFunc = func(opCtx *OpCtx) error {
		return nil
	}
	admin.downloadPkgFunc = func(opCtx *OpCtx, param DownloadParam) error {
		return nil
	}
	err := admin.StartAgent()
	if err != nil {
		t.Error(err)
		return
	}
	token := "test-token"
	command.NewFileTaskStore(testutil.TaskStoreDir).CreateStatus(command.ExecutionTokenFromString(token), command.StoredStatus{
		ResponseType: command.TypeStructured,
	})
	err = admin.ReinstallAgent(ReinstallParam{
		TaskToken: TaskToken{
			TaskToken: token,
		},
	})
	if err != nil {
		t.Error(err)
	}
}
