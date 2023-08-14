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

package host

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/oceanbase/obagent/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/monitor/plugins/common"
)

const (
	NetSnmp   = "/net/snmp"
	DiskStats = "/diskstats"
	NetProc   = "/proc"

	EnvDiskstats = "PROC_DISKSTATS"
	EnvSnmp      = "PROC_NET_SNMP"
	EnvRoot      = "PROC_ROOT"
)

var (
	zeroByte    = []byte("0")
	newLineByte = []byte("\n")
	colonByte   = []byte(":")
)

func gatherTcpretransRadio(ctx context.Context) (int64, int64, error) {
	var tcpOutSegs, tcpRetrans int64
	procNetSnmp := proc(EnvSnmp, NetSnmp)
	snmp, err := os.ReadFile(procNetSnmp)
	if err != nil {
		return tcpOutSegs, tcpRetrans, err
	}
	metricsMap := loadSnmpTable(snmp)
	if v, ok := metricsMap["TcpOutSegs"]; ok {
		tcpOutSegs, ok = v.(int64)
		if !ok {
			log.WithContext(ctx).Warn("TcpOutSegs is not expect type")
		}
	}
	if v, ok := metricsMap["TcpRetransSegs"]; ok {
		tcpRetrans, ok = v.(int64)
		if !ok {
			log.WithContext(ctx).Warn("TcpRetransSegs is not expect type")
		}
	}
	return tcpOutSegs, tcpRetrans, nil
}

// gather compute io_util and io_await need data
func gatherIoTicks(ctx context.Context) (map[string]float64, map[string]map[string]float64, error) {
	devIoTicks := make(map[string]float64)
	devIoAwaits := make(map[string]map[string]float64)
	procDiskstats := proc(EnvDiskstats, DiskStats)
	diskstats, err := os.ReadFile(procDiskstats)
	if err != nil {
		return devIoTicks, devIoAwaits, err
	}
	diskstatsMap := loadDiskstatsTable(diskstats)
	for diskName, stats := range diskstatsMap {
		ioTime, _ := stats["ioTicks"].(int64)
		devIoTicks[diskName] = float64(ioTime)

		tmp := make(map[string]float64)
		rdIos, _ := stats["readCount"].(int64)
		wrIos, _ := stats["writeCount"].(int64)
		rdTicks, _ := stats["readTicks"].(int64)
		wrTicks, _ := stats["writeTicks"].(int64)

		tmp["readCount"] = float64(rdIos)
		tmp["writeCount"] = float64(wrIos)
		tmp["readTicks"] = float64(rdTicks)
		tmp["writeTicks"] = float64(wrTicks)
		devIoAwaits[diskName] = tmp
	}
	return devIoTicks, devIoAwaits, err
}

func loadSnmpTable(table []byte) map[string]interface{} {
	entries := map[string]interface{}{}
	lines := bytes.Split(table, newLineByte)
	var value int64
	var err error
	for i := 0; i < len(lines); i = i + 2 {
		if len(lines[i]) == 0 {
			continue
		}
		headers := bytes.Fields(lines[i])
		prefix := bytes.TrimSuffix(headers[0], colonByte)
		metrics := bytes.Fields(lines[i+1])

		for j := 1; j < len(headers); j++ {
			// counter is zero
			if bytes.Equal(metrics[j], zeroByte) {
				entries[string(append(prefix, headers[j]...))] = int64(0)
				continue
			}
			// the counter is not zero
			value, err = strconv.ParseInt(string(metrics[j]), 10, 64)
			if err == nil {
				entries[string(append(prefix, headers[j]...))] = value
			}
		}
	}
	return entries
}

func loadDiskstatsTable(table []byte) map[string]map[string]interface{} {
	var diskstatsMap = make(map[string]map[string]interface{})
	lines := bytes.Split(table, newLineByte)
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) == 0 {
			continue
		}
		headers := bytes.Fields(lines[i])
		if len(headers) < 14 {
			log.Warnf("column of diskstats line %d is not correct!", i)
			continue
		}

		devMap := make(map[string]interface{})
		// ticks is equal to milliseconds
		devMap["readCount"], _ = strconv.ParseInt(string(headers[3]), 10, 64)
		devMap["mergedReadCount"], _ = strconv.ParseInt(string(headers[4]), 10, 64)
		devMap["readSectorsCount"], _ = strconv.ParseInt(string(headers[5]), 10, 64)
		devMap["readTicks"], _ = strconv.ParseInt(string(headers[6]), 10, 64)
		devMap["writeCount"], _ = strconv.ParseInt(string(headers[7]), 10, 64)
		devMap["mergedWriteCount"], _ = strconv.ParseInt(string(headers[8]), 10, 64)
		devMap["writeSectorsCount"], _ = strconv.ParseInt(string(headers[9]), 10, 64)
		devMap["writeTicks"], _ = strconv.ParseInt(string(headers[10]), 10, 64)
		devMap["iopsInProgress"], _ = strconv.ParseInt(string(headers[11]), 10, 64)
		devMap["ioTicks"], _ = strconv.ParseInt(string(headers[12]), 10, 64)
		devMap["aveqIoTicks"], _ = strconv.ParseInt(string(headers[13]), 10, 64)

		dev := string(headers[2])
		diskstatsMap[dev] = devMap
	}
	return diskstatsMap
}

// proc can be used to read file paths from env
func proc(env, path string) string {
	// try to read full file path
	if p := os.Getenv(env); p != "" {
		return p
	}
	// try to read root path or use default root path
	root := os.Getenv(EnvRoot)
	if root == "" {
		root = NetProc
	}
	return root + path
}

func parseEthtoolResult(output string) float64 {
	var value float64
	tags := strings.Split(output, "\n\t")
	for _, tag := range tags {
		if strings.HasPrefix(tag, "Speed:") {
			strs := strings.Split(tag, ": ")
			if len(strs) > 1 {
				speedStr := strings.Split(tag, ": ")[1]
				speed, err := strconv.Atoi(strings.Split(speedStr, "Mb/s")[0])
				if err == nil {
					value = float64(speed * 1000 * 1000)
					return value
				}
			}
		}
	}
	return value
}

func processChronycOutput(out string) (float64, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		stats := strings.Split(line, ":")
		if len(stats) < 2 {
			return 0, errors.Errorf("unexpected output from chronyc, expected ':' in %s", out)
		}
		name := strings.ToLower(strings.Replace(strings.TrimSpace(stats[0]), " ", "_", -1))

		valueFields := strings.Fields(stats[1])
		if len(valueFields) == 0 {
			return 0, errors.Errorf("unexpected output from chronyc: %s", out)
		}
		if strings.Contains(strings.ToLower(name), "last_offset") {
			value, err := strconv.ParseFloat(valueFields[0], 64)
			if err != nil {
				return 0, errors.Wrap(err, "parse last_offset from chronyc output")
			}
			return value, nil
		}
	}

	return 0, errors.Errorf("can not found last_offset in chronyc output")
}

func processNtpqOutput(out string) (float64, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "*") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return 0, errors.Errorf("ntpq output line is invalid, line: %s", line)
			}
			offsetMs, err := strconv.ParseFloat(fields[len(fields)-2], 64)
			if err != nil {
				return 0, err
			}
			return offsetMs / 1000, nil
		}
	}
	return 0, errors.Errorf("can not found ntp_server")
}

func checkNtpProcess(name string, ctx context.Context) bool {
	allProcesses := common.GetProcesses()
	for _, process := range allProcesses.Processes {
		if name == process.Name {
			return true
		}
	}
	return false
}
