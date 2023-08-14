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
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/shell"
)

const (
	getKernelParamCommand        = "sysctl -n %s"
	memInfoFile           string = "/proc/meminfo"

	checkLibExistsCmd    string = "ldconfig -p | grep %s | wc -l"
	checkPortOccupiedCmd string = "netstat -ant| awk '{ print $4 }'| grep \":${HOST_PORT}\" | wc -l"
)

type System interface {
	// CommandExists Check whether the specified command exists
	CommandExists(cmd string) bool

	// LibExists Check whether the specified lib exists
	LibExists(libName string) (bool, error)

	// UserExists Check whether the specified user exists
	UserExists(userName string) bool

	// GetHostInfo Get basic information of host
	GetHostInfo() (*HostInfo, error)

	// GetLocalIpAddress Get the network address of the host.
	// Equivalent to `hostname -i`
	GetLocalIpAddress() (string, error)

	// GetKernelParam Get kernel param
	GetKernelParam(kernelParamName string) (*KernelParam, error)

	// GetMemoryInfo Get memory info
	GetMemoryInfo() (*MemoryInfo, error)

	// CheckPortOccupied Check if the host port is occupied
	CheckPortOccupied(port int) (*PortOccupied, error)
}

var (
	kernelParamReg         = regexp.MustCompile(`^\w+[\w.]*\w$`)
	checkLibExistsParamReg = regexp.MustCompile(`^\w+$`)
)

type SystemImpl struct {
}

type HostInfo struct {
	HostName          string `json:"hostName"`          // host name, equivalent to result of `hostname`
	BootTime          uint64 `json:"bootTime"`          // boot time
	Uptime            uint64 `json:"upTime"`            // running time since boot time
	Os                string `json:"os"`                // operating system, e.g. linux
	OsPlatform        string `json:"osPlatform"`        // operation system platform, e.g. ubuntu, linuxmint
	OsPlatformFamily  string `json:"osPlatformFamily"`  // operation system platform family, e.g. debian, rhel
	OsPlatformVersion string `json:"osPlatformVersion"` // release version of operating system, e.g. 3.10.0-327.ali2010.alios7.x86_64
	KernelVersion     string `json:"kernelVersion"`     // version of the OS kernel (if available)
	Architecture      string `json:"architecture"`      // hardware platform, x86_64 or aarch64, equivalent to result of `uname -i`
	CpuModelName      string `json:"cpuModelName"`
	CpuCount          uint64 `json:"cpuCount"`
	TotalMemory       uint64 `json:"totalMemory"`
}

type PortOccupied struct {
	Port     int  `json:"port"`
	Occupied bool `json:"occupied"`
}

type CheckPortOccupiedResult struct {
	ResultList []PortOccupied `json:"resultList"`
}

type KernelParam struct {
	ParamKey   string   `json:"paramKey"`   // ParamKey 内核参数名称，如 net.core.somaxconn
	ParamValue []string `json:"paramValue"` // ParamValue 内核参数值
}

type MemoryInfo struct {
	InfoPairs []*MemoryInfoPair `json:"infoPairs"` // 内存信息对
}

type MemoryInfoPair struct {
	Key   string `json:"key"`   // Key 参数名称，如 MemTotal
	Value uint64 `json:"value"` // Value 参数值
}

func (s SystemImpl) CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func validateCheckLibExistsParam(paramName string) bool {
	if !checkLibExistsParamReg.MatchString(paramName) {
		return false
	}
	return true
}

func parseLibExistsInfo(output string) (int64, error) {
	output = strings.TrimSuffix(output, "\n")

	lineCount, err := strconv.ParseInt(output, 10, 32)
	if err != nil {
		return 0, err
	}
	return lineCount, nil
}

func (s SystemImpl) LibExists(libName string) (bool, error) {
	if !validateCheckLibExistsParam(libName) {
		return false, errors.New("invalid param")
	}

	executeResult, err := libShell.NewCommand(fmt.Sprintf(checkLibExistsCmd, libName)).WithUser(shell.RootUser).WithOutputType(shell.StdOutput).Execute()
	if err != nil {
		return false, errors.Wrapf(err, "get lib info failed")
	}

	lineCount, err := parseLibExistsInfo(executeResult.Output)
	if err != nil {
		return false, err
	}

	return lineCount > 0, nil
}

func (s SystemImpl) UserExists(userName string) bool {
	_, err := user.Lookup(userName)
	return err == nil
}

func (s SystemImpl) GetHostInfo() (*HostInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, errors.Errorf("failed to get host info: %s", err)
	}
	cpuInfos, err := cpu.Info()
	if err != nil {
		return nil, errors.Errorf("failed to get cpu info: %s", err)
	}
	count, err := cpu.Counts(true)
	if err != nil {
		return nil, errors.Errorf("failed to get cpu count: %s", err)
	}
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, errors.Errorf("failed to get memory info: %s", err)
	}
	return &HostInfo{
		HostName:          info.Hostname,
		BootTime:          info.BootTime,
		Uptime:            info.Uptime,
		Os:                info.OS,
		OsPlatform:        info.Platform,
		OsPlatformFamily:  info.PlatformFamily,
		OsPlatformVersion: info.PlatformVersion,
		KernelVersion:     info.KernelVersion,
		Architecture:      info.KernelArch,
		CpuModelName:      cpuInfos[0].ModelName,
		CpuCount:          uint64(count),
		TotalMemory:       memInfo.Total,
	}, nil
}

func (s SystemImpl) GetLocalIpAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.Errorf("failed to get local ip address, cannot get network interfaces: %s", err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("failed to get local ip address, no ip address found")
}

func validateKernelParamName(paramName string) bool {
	if !kernelParamReg.MatchString(paramName) {
		return false
	}
	return true
}

func parseKernelParam(paramName, cmdRet string) *KernelParam {
	paramValueSplits := strings.Fields(cmdRet)
	log.Debugf("get kernelParam %s, value:%v", paramName, cmdRet)

	return &KernelParam{
		ParamKey:   paramName,
		ParamValue: paramValueSplits,
	}
}

func (s SystemImpl) GetKernelParam(paramName string) (*KernelParam, error) {
	if !validateKernelParamName(paramName) {
		return nil, errors.Errorf("invalid paramName %s", paramName)
	}
	cmd := fmt.Sprintf(getKernelParamCommand, paramName)
	executeResult, err := libShell.NewCommand(cmd).WithUser(shell.RootUser).WithTimeout(5 * time.Second).WithOutputType(shell.StdOutput).Execute()
	if err != nil {
		return nil, errors.Wrapf(err, "get kernel param %s failed", paramName)
	}

	return parseKernelParam(paramName, executeResult.Output), nil
}

func (s SystemImpl) GetMemoryInfo() (*MemoryInfo, error) {
	exists, err := libFile.FileExists(memInfoFile)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("mem info not exists")
	}

	fd, err := os.Open(memInfoFile)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return loadMemInfoFromReader(fd)
}

func (s SystemImpl) CheckPortOccupied(port int) (*PortOccupied, error) {
	cmd := strings.ReplaceAll(checkPortOccupiedCmd, "${HOST_PORT}", strconv.Itoa(port))
	output, err := libShell.NewCommand(cmd).WithOutputType(shell.StdOutput).WithUser(shell.RootUser).WithTimeout(3 * time.Second).Execute()
	if err != nil {
		return nil, err
	}
	cnt, err := strconv.Atoi(strings.Trim(output.Output, "\n"))
	if err != nil {
		return nil, err
	}
	return &PortOccupied{Port: port, Occupied: cnt > 0}, nil
}

func loadMemInfoFromReader(fd io.Reader) (*MemoryInfo, error) {
	scanner := bufio.NewScanner(fd)
	memInfoPairs := make([]*MemoryInfoPair, 0)
	for scanner.Scan() {
		line := scanner.Text()
		memInfoPair, err := parseMemInfoPair(line)
		if err != nil {
			return nil, err
		}

		memInfoPairs = append(memInfoPairs, memInfoPair)
	}

	return &MemoryInfo{InfoPairs: memInfoPairs}, nil
}

func parseMemInfoPair(line string) (*MemoryInfoPair, error) {
	line = strings.TrimSuffix(line, "kB")
	lineSplits := strings.Split(line, ":")
	if len(lineSplits) != 2 {
		return nil, errors.New("invalid memory info format")
	}
	value := strings.TrimSpace(lineSplits[1])
	parsedValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, err
	}

	return &MemoryInfoPair{
		Key:   strings.TrimSpace(lineSplits[0]),
		Value: parsedValue,
	}, nil
}
