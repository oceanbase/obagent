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

package agentd

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	process2 "github.com/shirou/gopsutil/v3/process"

	"github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/tests/testutil"
)

var mockAgentPath string

func TestMain(m *testing.M) {
	testutil.MakeDirs()
	//err := testutil.BuildBins()
	//if err != nil {
	//	os.Exit(2)
	//}
	err := testutil.BuildMockAgent()
	if err != nil {
		os.Exit(2)
	}
	ret := m.Run()
	testutil.KillAll()
	testutil.DelTestFiles()
	os.Exit(ret)
}

func TestNewWatchdog(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()

	watchdog := NewAgentd(Config{
		RunDir:          testutil.RunDir,
		LogDir:          testutil.LogDir,
		LogLevel:        "info",
		CleanupDangling: true,
		Services: map[string]ServiceConfig{
			"test": {
				Program:        testutil.MockAgentPath,
				Args:           []string{"test", testutil.RunDir, "1", "1"},
				RunDir:         testutil.RunDir,
				FinalWait:      time.Second * 2,
				MinLiveTime:    time.Second * 5,
				QuickExitLimit: 3,
				Stdout:         testutil.LogDir + "/test.output.log",
				Stderr:         testutil.LogDir + "/test.error.log",
			},
		},
	})
	err := watchdog.Start()
	if err != nil {
		t.Error(err)
	}

	err = watchdog.Start()
	if err != nil {
		t.Error("duplicate start should not fail")
	}

	status := watchdog.Status()
	if status.State != http.Running {
		t.Error("watchdog state should be running")
	}

	time.Sleep(1500 * time.Millisecond)

	status = watchdog.Status()
	fmt.Printf("status: %+v\n", status)
	if status.Services["test"].State != http.Running {
		t.Errorf("service state should be running got '%s'", status.Services["test"].State)
	}
	err = watchdog.Stop()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
	for name, service := range watchdog.services {
		_, err2 := process2.NewProcess(int32(service.Pid()))
		if err2 == nil {
			t.Errorf("service '%s' should be stopped", name)
		}
	}
	err = watchdog.Stop()
	if err != nil {
		t.Error("duplicate stop should not fail")
	}
}

func TestWatchdog_ListenSignal(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()

	watchdog := NewAgentd(Config{
		RunDir:   testutil.RunDir,
		LogDir:   testutil.LogDir,
		LogLevel: "info",
		Services: map[string]ServiceConfig{
			"test": {
				Program:        testutil.MockAgentPath,
				Args:           []string{"test", testutil.RunDir, "1", "1"},
				RunDir:         testutil.RunDir,
				FinalWait:      time.Second * 2,
				MinLiveTime:    time.Second * 5,
				QuickExitLimit: 3,
				Stdout:         testutil.LogDir + "/test.output.log",
				Stderr:         testutil.LogDir + "/test.error.log",
			},
		},
	})
	err := watchdog.Start()
	if err != nil {
		t.Error(err)
	}

	p, err := process2.NewProcess(int32(watchdog.Status().Pid))
	if err != nil {
		t.Error(err)
		return
	}
	ch := make(chan struct{})
	go func() {
		watchdog.ListenSignal()
		close(ch)
	}()
	time.Sleep(100 * time.Millisecond)
	_ = p.SendSignal(syscall.SIGTERM)
	select {
	case <-ch:
		// ok
	case <-time.After(time.Second * 5):
		t.Error("wait stop by signal timeout")
	}
	status := watchdog.Status()
	if status.State != http.Stopped {
		t.Error("watchdog should be stopped")
	}

}

func TestStartFail(t *testing.T) {
	defer testutil.KillAll()
	testutil.MakeDirs()
	defer testutil.DelTestFiles()

	watchdog := NewAgentd(Config{
		RunDir:   testutil.RunDir,
		LogDir:   testutil.LogDir,
		LogLevel: "info",
		Services: map[string]ServiceConfig{
			"test": {
				Program:        testutil.MockAgentPath,
				Args:           []string{"test", testutil.RunDir, "-1", "1"},
				RunDir:         testutil.RunDir,
				FinalWait:      time.Second * 2,
				MinLiveTime:    time.Second * 5,
				QuickExitLimit: 0,
				Stdout:         testutil.LogDir + "/test.output.log",
				Stderr:         testutil.LogDir + "/test.error.log",
			},
		},
	})
	err := watchdog.Start()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(2000 * time.Millisecond)
	if watchdog.services["test"].State().State != http.Stopped {
		t.Error("failed service should be 'stopped'")
	}
}
