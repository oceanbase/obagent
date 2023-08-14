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

package system

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/oceanbase/obagent/lib/shell"
)

const GetProcessTcpListenCommand = "netstat -tunlp 2>/dev/null | { grep '%d/' || true; }"

const ListProcessProcInfoByName = "pgrep %s$ | xargs -I {} ls -l /proc/{}/%s"

const ListProcessProcInfoByPid = "ls -l /proc/%d/%s"

var libShell shell.Shell = shell.ShellImpl{}

type Process interface {
	// ListAllPids Returns PIDs of all processes
	ListAllPids() ([]int32, error)

	// GetProcessInfoByPid Get process info by PID
	GetProcessInfoByPid(pid int32) (*ProcessInfo, error)

	// ListProcesses List all processes
	ListProcesses() ([]*ProcessInfo, error)

	// FindProcessInfoByName Find process info by process name
	FindProcessInfoByName(name string) ([]*ProcessInfo, error)

	// FindProcessInfoByKeyword Find process info by process keyword
	FindProcessInfoByKeyword(keyword string) ([]*ProcessInfo, error)

	// ProcessExists Check whether process exists by process name
	ProcessExists(name string) (bool, error)

	// TerminateProcessByPid Terminate process by PID, equivalent to `kill` or `kill -15`
	TerminateProcessByPid(pid int32) error

	// TerminateProcessByName Terminate process by process name, equivalent to `kill` or `kill -15`
	TerminateProcessByName(name string) error

	// TerminateProcessByKeyword Terminate process by process keyword, equivalent to `kill` or `kill -15`
	TerminateProcessByKeyword(keyword string) error

	// KillProcessByPid Kill process by PID, equivalent to `kill -9`
	KillProcessByPid(pid int32) error

	// KillProcessByName Kill process by process name, equivalent to `kill -9`
	KillProcessByName(name string) error

	// KillProcessByKeyword Kill process by process keyword, equivalent to `kill -9`
	KillProcessByKeyword(keyword string) error

	// GetProcessProcInfoByPid ls -l /proc/{pid}/[infoType]
	GetProcessProcInfoByPid(pid int32, infoType string, user string) ([]string, error)

	// GetProcessProcInfoByName pgrep %s$ | xargs -I {} ls -l /proc/{}/%s
	GetProcessProcInfoByName(name string, infoType string, user string) ([]string, error)
}

type ProcessImpl struct {
}

type ProcessInfo struct {
	Pid               int32     `json:"pid"`
	Name              string    `json:"name"`              // process name
	StartCommand      string    `json:"startCommand"`      // process command line, with arguments
	Username          string    `json:"username"`          // username of the process
	Ports             []int     `json:"ports"`             // the host ports the process occupied
	CreateTime        time.Time `json:"createTime"`        // process create time
	ElapsedTimeMillis int64     `json:"elapsedTimeMillis"` // elapsed time since process created
}

func (p ProcessInfo) String() string {
	return fmt.Sprintf("ProcessInfo{pid: %v, name: %v, cmdLine: %v, username: %v, createTime: %v, elapsedTime: %v}", p.Pid, p.Name, p.StartCommand, p.Username, p.CreateTime, p.ElapsedTimeMillis)
}

func (p ProcessImpl) ListAllPids() ([]int32, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, errors.Errorf("failed to list pids: %s", err)
	}
	return pids, nil
}

func (p ProcessImpl) GetProcessProcInfoByName(name string, infoType string, user string) ([]string, error) {
	cmd := fmt.Sprintf(ListProcessProcInfoByName, name, infoType)
	return execListProcessProcCmd(cmd, user)
}

func (p ProcessImpl) GetProcessProcInfoByPid(pid int32, infoType string, user string) ([]string, error) {
	cmd := fmt.Sprintf(ListProcessProcInfoByPid, pid, infoType)
	return execListProcessProcCmd(cmd, user)
}

func (p ProcessImpl) GetProcessInfoByPid(pid int32) (*ProcessInfo, error) {
	proc, err := p.getProcess(pid)
	if err != nil {
		return nil, errors.Wrapf(err, "get process info by pid %d", pid)
	}
	return processToProcessInfo(proc), nil
}

func (p ProcessImpl) ListProcesses() ([]*ProcessInfo, error) {
	processes, err := p.listProcesses()
	if err != nil {
		return nil, errors.Wrapf(err, "list processes")
	}
	result := make([]*ProcessInfo, 0, len(processes))
	for _, proc := range processes {
		result = append(result, processToProcessInfo(proc))
	}
	return result, nil
}

func (p ProcessImpl) FindProcessInfoByName(name string) ([]*ProcessInfo, error) {
	processes, err := p.findProcessByName(name)
	if err != nil {
		return nil, errors.Wrapf(err, "find process info by name %s", name)
	}
	result := make([]*ProcessInfo, 0, len(processes))
	for _, proc := range processes {
		result = append(result, processToProcessInfo(proc))
	}
	return result, nil
}

func (p ProcessImpl) FindProcessInfoByKeyword(keyword string) ([]*ProcessInfo, error) {
	processes, err := p.findProcessByKeyword(keyword)
	if err != nil {
		return nil, errors.Wrapf(err, "find process info by keyword %s", keyword)
	}
	result := make([]*ProcessInfo, 0)
	for _, proc := range processes {
		result = append(result, processToProcessInfo(proc))
	}
	return result, nil
}

func (p ProcessImpl) ProcessExists(name string) (bool, error) {
	processes, err := p.findProcessByName(name)
	if err != nil {
		return false, errors.Wrapf(err, "check process exists by name %s", name)
	}
	return len(processes) > 0, nil
}

func (p ProcessImpl) TerminateProcessByPid(pid int32) error {
	proc, err := p.getProcess(pid)
	if err != nil {
		return errors.Wrapf(err, "terminate process by pid %d", pid)
	}
	err = p.terminateProcess(proc)
	if err != nil {
		return errors.Wrapf(err, "terminate process by pid %d", pid)
	}
	return nil
}

func (p ProcessImpl) TerminateProcessByName(name string) error {
	processes, err := p.findProcessByName(name)
	if err != nil {
		return errors.Wrapf(err, "terminate process by name %s", name)
	}
	for _, proc := range processes {
		err := p.terminateProcess(proc)
		if err != nil {
			return errors.Wrapf(err, "terminate process by name %s", name)
		}
	}
	return nil
}

func (p ProcessImpl) TerminateProcessByKeyword(keyword string) error {
	processes, err := p.findProcessByKeyword(keyword)
	if err != nil {
		return errors.Wrapf(err, "terminate process by keyword %s", keyword)
	}
	for _, proc := range processes {
		err := p.terminateProcess(proc)
		if err != nil {
			return errors.Wrapf(err, "terminate process by keyword %s", keyword)
		}
	}
	return nil
}

func (p ProcessImpl) KillProcessByPid(pid int32) error {
	proc, err := p.getProcess(pid)
	if err != nil {
		return errors.Wrapf(err, "kill process by pid %d", pid)
	}
	err = p.killProcess(proc)
	if err != nil {
		return errors.Wrapf(err, "kill process by pid %d", pid)
	}
	return nil
}

func (p ProcessImpl) KillProcessByName(name string) error {
	processes, err := p.findProcessByName(name)
	if err != nil {
		return errors.Wrapf(err, "kill process by name %s", name)
	}
	for _, proc := range processes {
		err := p.killProcess(proc)
		if err != nil {
			return errors.Wrapf(err, "kill process by name %s", name)
		}
	}
	return nil
}

func (p ProcessImpl) KillProcessByKeyword(keyword string) error {
	processes, err := p.findProcessByKeyword(keyword)
	if err != nil {
		return errors.Wrapf(err, "kill process by keyword %s", keyword)
	}
	for _, proc := range processes {
		err := p.killProcess(proc)
		if err != nil {
			return errors.Wrapf(err, "kill process by keyword %s", keyword)
		}
	}
	return nil
}

func (p ProcessImpl) getProcess(pid int32) (*process.Process, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, errors.Errorf("failed to get process from pid %d: %s", pid, err)
	}
	return proc, nil
}

func (p ProcessImpl) listProcesses() ([]*process.Process, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, errors.Errorf("failed to list processes: %s", err)
	}
	var ret []*process.Process
	for _, pid := range pids {
		proc := &process.Process{
			Pid: pid,
		}
		ret = append(ret, proc)
	}
	return ret, nil
}

func (p ProcessImpl) findProcessByName(name string) ([]*process.Process, error) {
	processes, err := p.listProcesses()
	if err != nil {
		return nil, errors.Wrapf(err, "find process by name %s", name)
	}
	result := make([]*process.Process, 0)
	for _, proc := range processes {
		if procName, err := proc.Name(); err == nil && procName == name {
			result = append(result, proc)
		}
	}
	return result, nil
}

func (p ProcessImpl) findProcessByKeyword(keyword string) ([]*process.Process, error) {
	processes, err := p.listProcesses()
	if err != nil {
		return nil, errors.Wrapf(err, "find process by keyword %s", keyword)
	}
	result := make([]*process.Process, 0)
	for _, proc := range processes {
		if cmdline, err := proc.Cmdline(); err == nil && strings.Contains(cmdline, keyword) {
			result = append(result, proc)
		}
	}
	return result, nil
}

func (p ProcessImpl) terminateProcess(proc *process.Process) error {
	err := proc.Terminate()
	if err != nil {
		return errors.Errorf("failed to terminate process with pid %d: %s", proc.Pid, err)
	}
	return nil
}

func (p ProcessImpl) killProcess(proc *process.Process) error {
	err := proc.Kill()
	if err != nil {
		return errors.Errorf("failed to kill process with pid %d: %s", proc.Pid, err)
	}
	return nil
}

func processToProcessInfo(p *process.Process) *ProcessInfo {
	pi := ProcessInfo{
		Ports: []int{},
	}
	pi.Pid = p.Pid
	if createTimeMillis, err := p.CreateTime(); err == nil {
		createTime := time.Unix(createTimeMillis/1000, (createTimeMillis%1000)*1000)
		pi.CreateTime = createTime
		elapsedTime := time.Now().Sub(createTime)
		pi.ElapsedTimeMillis = int64(elapsedTime / time.Millisecond)
	}
	if name, err := p.Name(); err == nil {
		pi.Name = name
	}
	if cmdline, err := p.Cmdline(); err == nil {
		pi.StartCommand = cmdline
	}
	if username, err := p.Username(); err == nil {
		pi.Username = username
	}
	if ports, err := getProcessOccupiedPorts(p.Pid, pi.Username); err == nil {
		pi.Ports = ports
	}
	return &pi
}

func getProcessOccupiedPorts(pid int32, user string) ([]int, error) {
	cmd := fmt.Sprintf(GetProcessTcpListenCommand, pid)
	executeResult, err := libShell.NewCommand(cmd).WithUser(user).WithOutputType(shell.StdOutput).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "get process occupied ports")
	}
	output := strings.TrimSpace(executeResult.Output)
	if output == "" {
		return []int{}, nil
	}
	occupiedPorts := make([]int, 0)
	tcpLines := strings.Split(output, "\n")
	for _, tcpLine := range tcpLines {
		port, ok := parseNetstatLine(tcpLine)
		if !ok {
			return nil, errors.Errorf("failed to get process occupied ports, invalid output line: %s", tcpLine)
		}
		occupiedPorts = append(occupiedPorts, port)
	}
	// One occupied port can corresponds to multiple netstat lines, so remove duplicate ports.
	return removeDuplicate(occupiedPorts), nil
}

func removeDuplicate(ports []int) []int {
	if len(ports) == 0 {
		return []int{}
	}
	result := make([]int, 0)
	var seen = make(map[int]bool, len(ports))
	for _, port := range ports {
		if _, exists := seen[port]; !exists {
			result = append(result, port)
			seen[port] = true
		}
	}
	return result
}

func parseNetstatLine(line string) (int, bool) {
	fields := strings.Fields(line)
	if len(fields) != 7 {
		return 0, false
	}
	field := fields[3]
	if len(field) == 0 || !strings.Contains(field, ":") {
		return 0, false
	}
	i := strings.LastIndex(field, ":")
	if i == -1 {
		return 0, false
	}
	portStr := field[i+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, false
	}
	return port, true
}

func execListProcessProcCmd(cmd string, user string) ([]string, error) {
	executeResult, err := libShell.NewCommand(cmd).WithUser(user).WithOutputType(shell.StdOutput).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "get process proc info")
	}
	infos := make([]string, 0)
	output := strings.TrimSpace(executeResult.Output)

	if output == "" {
		return infos, nil
	}

	lines := strings.Split(output, "\n")
	return lines, nil
}
