package process

import (
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestProc(t *testing.T) {
	proc := NewProc(ProcessConfig{
		Program: "echo",
		Args:    []string{"hello"},
	})

	err := proc.Start(nil)
	if err != nil {
		t.Error("start process failed", err)
		return
	}

	state := proc.State()
	if state.Pid <= 0 {
		t.Error("bad state pid")
	}
	if state.Exited == true {
		t.Error("bad state exited")
	}
	state = proc.Wait()
	state = proc.State()
	if state.Pid <= 0 {
		t.Error("bad state pid")
	}
	if state.Exited == false {
		t.Error("bad state exited")
	}
	if state.Success == false {
		t.Error("bad state success")
	}
	if string(state.Stdout) != "hello\n" {
		t.Errorf("stdout wrong: '%s'", string(state.Stdout))
	}
	if string(state.Stderr) != "" {
		t.Errorf("stderr wrong, '%s'", string(state.Stderr))
	}
	select {
	case <-proc.Done():
	default:
		t.Error("process not done")
	}
}

func uniqName() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int())
}

func TestProc_Stop(t *testing.T) {
	proc := NewProc(ProcessConfig{
		Program:   "sleep",
		Args:      []string{"5"},
		KillWait:  1 * time.Second,
		FinalWait: 1 * time.Second,
	})
	err := proc.Stop()
	if err == nil {
		t.Error("should failed")
	}

	_ = proc.Start(nil)
	if !proc.State().Running {
		t.Error("process should running")
	}

	proc.Stop()
	time.Sleep(1 * time.Second)
	if proc.State().Running {
		t.Error("process should be stopped")
	}
}

func TestCmd(t *testing.T) {
	cmd, err := createCmd(ProcessConfig{
		Program: "sleep",
		Args:    []string{"1"},
	})
	if err != nil {
		t.Error(err)
	}
	if !strings.HasSuffix(cmd.Path, "/sleep") || !reflect.DeepEqual(cmd.Args, []string{"sleep", "1"}) {
		t.Error("bad command")
	}

	if cmd.Stdin != nil || cmd.Stdout != nil || cmd.Stderr != nil {
		t.Error("stdio should be nil")
	}
	if cmd.SysProcAttr.Credential != nil {
		t.Error("Credential should be nil")
	}
}

func TestCmdIO(t *testing.T) {
	outFile := os.TempDir() + "/out"
	errFile := os.TempDir() + "/errors"
	defer os.Remove(outFile)
	defer os.Remove(errFile)

	cmd, err := createCmd(ProcessConfig{
		Program: "sleep",
		Args:    []string{"1"},
		Stdout:  outFile,
		Stderr:  errFile,
	})
	if err != nil {
		t.Error("create cmd failed", err)
		return
	}
	if cmd.Stdin != nil || cmd.Stdout == nil || cmd.Stderr == nil {
		t.Error("stdout, stderr should be nil")
	}
}

func TestCmdUser(t *testing.T) {
	uid := os.Getuid()
	gid := os.Getgid()
	u, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		t.Errorf("can not find user of %d", uid)
		return
	}
	g, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		t.Errorf("can not find group of %d", gid)
		return
	}

	cmd, err := createCmd(ProcessConfig{
		Program: "sleep",
		Args:    []string{"1"},
		User:    u.Username,
		Group:   g.Name,
	})
	if err != nil {
		t.Error("create cmd failed", err)
		return
	}
	if cmd.SysProcAttr.Credential == nil {
		t.Error("Credential should not be nil")
	}
	if cmd.SysProcAttr.Credential.Uid != uint32(uid) || cmd.SysProcAttr.Credential.Gid != uint32(gid) {
		t.Errorf("bad uid or gid. expected: uid=%d,uname=%s gid=%d,gname=%s, got: uid:%d, gid:%d", uid, u.Name, gid, g.Name,
			cmd.SysProcAttr.Credential.Uid, cmd.SysProcAttr.Credential.Gid)
	}
	_, err = createCmd(ProcessConfig{
		Program: "sleep",
		Args:    []string{"1"},
		User:    "no_such_user",
	})
	if err == nil {
		t.Error("should be error")
	}
	_, err = createCmd(ProcessConfig{
		Program: "sleep",
		Args:    []string{"1"},
		Group:   "no_such_group",
	})
	if err == nil {
		t.Error("should be error")
	}
}

func TestCmdEnv(t *testing.T) {
	cmd, err := createCmd(ProcessConfig{
		Program: "sleep",
		Args:    []string{"1"},
		Envs: map[string]string{
			"NEW_ENV": "NEW_ENV_VALUE",
		},
	})
	if err != nil {
		t.Error("create process failed")
		return
	}
	for _, e := range cmd.Env {
		if e == "NEW_ENV=NEW_ENV_VALUE" {
			return
		}
	}
	t.Error("env not added")
}

func TestCmdRlimit(t *testing.T) {
	cmd, err := createCmd(ProcessConfig{
		Program: "sleep",
		Args:    []string{"1"},
		Rlimit: map[string]int64{
			RLimitProcess: 1000,
		},
	})
	if err != nil {
		t.Error("create process failed")
		return
	}
	if !strings.HasSuffix(cmd.Path, "/sh") {
		t.Error("not using sh")
	}
	if !strings.Contains(cmd.Args[2], " -u 1000") {
		t.Error("bad ulimit")
	}

	_, err = createCmd(ProcessConfig{
		Program: "sleep",
		Args:    []string{"1"},
		Rlimit: map[string]int64{
			"xxx": 1000,
		},
	})
	if err == nil {
		t.Error("bad rlimit should trigger error")
	}
}

func Test_replaceArgs(t *testing.T) {
	args := []string{"hello ${name}"}
	ret := replaceArgs(args, nil)
	if !reflect.DeepEqual(args, ret) {
		t.Error("should not replace")
	}
	ret = replaceArgs(args, replacer(map[string]string{"name": "world"}))
	if !reflect.DeepEqual(ret, []string{"hello world"}) {
		t.Error("should replaced")
	}
}

func TestRun(t *testing.T) {
	state, err := Run("echo", "hello")
	if err != nil {
		t.Error(err.Error())
	}
	if string(state.Stdout) != "hello\n" {
		t.Error("run result wrong")
	}
}
