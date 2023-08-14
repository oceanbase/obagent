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

package agentd

import (
	"fmt"
	"os"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/agentd/api"
	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/lib/process"
)

// serviceProc creates a proc via ServiceConfig
func serviceProc(conf ServiceConfig) *process.Proc {
	return process.NewProc(toProcConfig(conf))
}

func toProcConfig(conf ServiceConfig) process.ProcessConfig {
	if conf.Cwd == "" {
		conf.Cwd = path.AgentDir()
	}
	return process.ProcessConfig{
		Program:    conf.Program,
		Args:       conf.Args,
		Stdout:     conf.Stdout,
		Stderr:     conf.Stderr,
		Cwd:        conf.Cwd,
		InheritEnv: true,
		KillWait:   conf.KillWait,
		FinalWait:  conf.FinalWait,
	}
}

type Service struct {
	name    string
	conf    ServiceConfig
	proc    *process.Proc
	done    chan struct{}
	limiter Limiter
	state   *http.StateHolder
}

type TaskParam struct {
	Action string `json:"action"`
}

func NewService(name string, conf ServiceConfig) *Service {
	proc := serviceProc(conf)
	limiter, err := NewLimiter(name, conf.Limit)
	if err != nil {
		log.WithError(err).Errorf("failed to create limiter for service '%s' with config %+v", name, conf.Limit)
		limiter = &NopLimiter{}
	}
	ret := &Service{
		name:    name,
		conf:    conf,
		proc:    proc,
		done:    nil,
		limiter: limiter,
		state:   http.NewStateHolder(http.Stopped),
	}
	return ret
}

func timeToMill(t time.Time) int64 {
	ret := t.UnixNano() / time.Millisecond.Nanoseconds()
	if ret > 0 {
		return ret
	}
	return 0
}

type ServiceState struct {
	//Running whether the process is running
	Running bool `json:"running"`
	//Success whether the process is exited with code 0
	Success bool `json:"success"`
	//Exited whether the process is exited
	Exited bool `json:"exited"`

	//Pid of the process
	Pid int `json:"Pid"`
	//ExitCode of the finished process
	ExitCode int `json:"exit_code"`

	//StarAt time of the process started
	StartAt int64 `json:"start_at"`
	//EndAt time of the process ended
	EndAt int64 `json:"end_at"`

	Status string `json:"status"`
}

func (s *Service) Start() (err error) {
	s.cleanup()
	if s.state.Get() == http.Running {
		return nil
	}
	if !s.state.Cas(http.Stopped, http.Starting) {
		err = ServiceAlreadyStartedErr.NewError(s.name)
		log.WithField("service", s.name).WithError(err).Warn("service already started")
		return err
	}
	s.done = make(chan struct{})
	err = s.startProc()
	if err != nil {
		s.state.Set(http.Stopped)
		return
	}
	go s.guard()
	return nil
}

func (s *Service) startProc() (err error) {
	s.cleanup()
	pidFile, err := createPid(s.pidPath())
	if err != nil {
		return err
	}
	defer pidFile.Close()
	params := map[string]string{
		"run_dir":   s.conf.RunDir,
		"conf_dir":  path.ConfDir(),
		"bin_dir":   path.BinDir(),
		"agent_dir": path.AgentDir(),
	}
	log.WithField("service", s.name).Infof("starting service")
	err = s.proc.Start(params)
	if err != nil {
		err = InternalServiceErr.NewError(s.name).WithCause(err)
		s.cleanup()
		return err
	}
	log.WithField("service", s.name).Infof("service process started. pid: %d", s.Pid())
	err = writePid(pidFile, s.Pid())
	if err != nil {
		log.WithField("service", s.name).WithError(err).Errorf("write pid file failed %s", s.pidPath())
	}
	return
}

func (s *Service) waitExit() {
	select {
	case <-s.proc.Done():
	case <-time.After(time.Second):
	}
}

func (s *Service) limitResource() {
	err := s.limiter.LimitPid(s.Pid())
	if err != nil {
		log.WithField("service", s.name).WithError(err).Errorf("limit service resource failed. pid=%d", s.Pid())
	}
}

func (s *Service) guard() {
	var err error
	quickExitCount := 0
	defer func() {
		s.state.Set(http.Stopped)
		s.cleanup()
		close(s.done)
	}()
	s.limitResource()
	for {
		//todo wait for s.ready()
		s.state.Cas(http.Starting, http.Running)

		//s.queryStatus()
		s.waitExit()

		svcState := s.state.Get()
		state := s.proc.State()

		if state.Exited {
			log.WithField("service", s.name).Warnf("service exited with code %d. service state: %v", state.ExitCode, svcState)
		}

		// service is stopped by watchdog, exit guard
		if svcState == http.Stopping || svcState == http.Stopped {
			log.WithField("service", s.name).Infof("service stopped. service state: %s", svcState)
			return
		}

		if !state.Exited {
			// still running...
			continue
		}

		// process exited
		if state.ExitCode == 0 {
			log.WithField("service", s.name).Info("service normally exited")
			return
		}
		s.state.Set(http.Stopped)
		liveTime := state.EndAt.Sub(state.StartAt)
		if s.conf.MinLiveTime > 0 && liveTime < s.conf.MinLiveTime {
			quickExitCount++
			log.WithField("service", s.name).Warnf("service exited too quickly. live time: %d, MinLiveTime: %d, count: %d", liveTime, s.conf.MinLiveTime, quickExitCount)
			if quickExitCount >= s.conf.QuickExitLimit {
				log.WithField("service", s.name).Errorf("service exited too quickly. live time: %d, MinLiveTime: %d, count: %d", liveTime, s.conf.MinLiveTime, quickExitCount)
				return
			}
		} else {
			quickExitCount = 0
		}
		log.WithField("service", s.name).Info("recovering service")
		s.state.Set(http.Starting)
		err = s.startProc()
		if err != nil {
			s.state.Set(http.Stopped)
			log.WithField("service", s.name).WithError(err).Error("start service got error")
			return
		}
		s.limitResource()
	}
}

func (s *Service) Stop() (err error) {
	if s.state.Get() == http.Stopped {
		return nil
	}
	s.state.Set(http.Stopping) // state may in running, staring, stopping
	log.WithField("service", s.name).Info("stopping service")
	err = s.proc.Stop()
	if err != nil {
		err = InternalServiceErr.NewError(s.name).WithCause(err)
		log.WithField("service", s.name).WithField("pid", s.Pid()).WithError(err).Warn("stop service got error")
		state := s.State()
		if state.State == http.Stopping || state.State == http.Stopped {
			s.state.Set(state.State)
			return nil
		} else {
			log.WithField("service", s.name).WithField("pid", s.Pid()).Warn("service did not handle TERM signal properly, try KILL it")
			err = s.proc.Kill()
		}
	}
	return
}

func (s *Service) Pid() int {
	if s.proc == nil {
		return 0
	}
	return s.proc.Pid()
}

func (s *Service) cleanup() {
	socketPath := s.socketPath()
	if isSocket(socketPath) {
		//log.WithField("service", s.name).Info("removing socket file %s", socketPath)
		_ = os.Remove(socketPath)
	}
	_ = removePid(s.pidPath(), s.Pid())
	_ = removePid(s.backupPidPath(), s.Pid())
}

func (s *Service) socketPath() string {
	return agent.SocketPath(s.conf.RunDir, s.name, s.Pid())
}

func (s *Service) pidPath() string {
	return agent.PidPath(s.conf.RunDir, s.name)
}

func (s *Service) backupPidPath() string {
	return agent.BackupPidPath(s.conf.RunDir, s.name, s.Pid())
}

func (s *Service) State() api.ServiceStatus {
	state := s.proc.State()
	svcState := http.Unknown
	if state.Running {
		ret, err := s.queryStatus()
		if err == nil {
			return api.ServiceStatus{
				Status: ret,
				Socket: agent.SocketPath(s.conf.RunDir, s.name, s.Pid()),
				EndAt:  state.EndAt.UnixNano(),
			}
		}
	} else {
		svcState = http.Stopped
	}
	return api.ServiceStatus{
		Status: http.Status{
			State:   svcState,
			Pid:     state.Pid,
			StartAt: state.StartAt.UnixNano(),
		},
		Socket: agent.SocketPath(s.conf.RunDir, s.name, s.Pid()),
		EndAt:  state.EndAt.UnixNano(),
	}
}

func (s *Service) queryStatus() (http.Status, error) {
	readyResult := http.Status{}
	c := s.apiClient()
	if c == nil {
		return readyResult, http.NoApiClientErr.NewError()
	}
	err := c.Call("/api/v1/status", nil, &readyResult)
	if err != nil {
		return readyResult, err
	}
	return readyResult, nil
}

func (s *Service) apiClient() *http.ApiClient {
	socketPath := s.socketPath()
	if isSocket(socketPath) {
		return http.NewSocketApiClient(socketPath, time.Second*5)
	}
	return nil
}

func createPid(pidPath string) (*os.File, error) {
	ret, err := os.OpenFile(pidPath, os.O_RDWR|os.O_CREATE|os.O_EXCL|syscall.O_CLOEXEC, 0644)
	if err != nil {
		return nil, WritePidFailedErr.NewError(pidPath).WithCause(err)
	}
	return ret, err
}

func writePid(f *os.File, pid int) error {
	_, err := fmt.Fprintf(f, "%d\n", pid)
	if err != nil {
		return WritePidFailedErr.NewError(f.Name()).WithCause(err)
	}
	return nil
}

func removePid(pidPath string, expectedPid int) error {
	pid, err := agent.ReadPid(pidPath)
	if err != nil {
		return RemovePidFailedErr.NewError(pidPath).WithCause(err)
	}
	if pid != expectedPid {
		return nil
	}
	log.Infof("remove pid file %s", pidPath)
	err = os.Remove(pidPath)
	if err != nil {
		return RemovePidFailedErr.NewError(pidPath).WithCause(err)
	}
	return nil
}

func isSocket(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && (stat.Mode()&os.ModeSocket != 0)
}
