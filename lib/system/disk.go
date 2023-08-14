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
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/disk"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/file"
	"github.com/oceanbase/obagent/lib/shell"
)

const getMountPointsCommand = "mount -t ext4,xfs | awk '{print $1, $3, $5 ,$6}'"
const sysFsBaseDir = "/sys/block"
const sysFsDevDir = "/sys/block/%s"
const sysFsPartitionDir = "/sys/block/%s/%s"
const sysFsTotalSize = "cat /sys/block/%s/size"
const sysFsSectorSize = "cat /sys/block/%s/queue/hw_sector_size"

var libFile file.File = file.FileImpl{}

type Disk interface {
	// GetDiskUsage Returns disk usage stats on the file system containing the given path.
	// Equivalent to `df <path>` command.
	GetDiskUsage(path string) (*DiskUsage, error)
	// BatchGetDiskInfos GetDiskInfos Returns disk basic info.
	BatchGetDiskInfos() ([]*DiskInfo, error)
}

type DiskImpl struct {
}

type DiskUsage struct {
	Path                 string  `json:"path"`
	TotalSizeBytes       uint64  `json:"totalSizeBytes"`       // total size in bytes
	TotalSizeDisplay     string  `json:"totalSizeDisplay"`     // total size in human readable format, e.g. 537 GB
	UsedSizeBytes        uint64  `json:"usedSizeBytes"`        // used size in bytes
	UsedSizeDisplay      string  `json:"usedSizeDisplay"`      // used size in human readable format, e.g. 6.9 GB
	AvailableSizeBytes   uint64  `json:"availableSizeBytes"`   // available size in bytes
	AvailableSizeDisplay string  `json:"availableSizeDisplay"` // available size in human readable format, e.g. 203 GB
	UsedPercent          float64 `json:"usedPercent"`          // used percent, range: [0, 100]
	UsedPercentDisplay   string  `json:"usedPercentDisplay"`   // used percent in human readable format, e.g. 91%
}

type DiskInfo struct {
	Path       string `json:"path"`       // Path 请求路径
	TotalSize  uint64 `json:"totalSize"`  // TotalSize 对应磁盘容量
	SectorSize uint64 `json:"sectorSize"` // SectorSize 对应磁盘定义的扇区大小
	MountInfo
}

func (d DiskImpl) GetDiskUsage(path string) (*DiskUsage, error) {
	usageStat, err := disk.Usage(path)
	if err != nil {
		return nil, errors.Errorf("failed to get disk usage: %s", err)
	}

	diskInfo := &DiskUsage{
		Path:                 path,
		TotalSizeBytes:       usageStat.Total,
		TotalSizeDisplay:     humanize.Bytes(usageStat.Total),
		UsedSizeBytes:        usageStat.Used,
		UsedSizeDisplay:      humanize.Bytes(usageStat.Used),
		AvailableSizeBytes:   usageStat.Free,
		AvailableSizeDisplay: humanize.Bytes(usageStat.Free),
		UsedPercent:          usageStat.UsedPercent,
		UsedPercentDisplay:   fmt.Sprintf("%2.1f%%", usageStat.UsedPercent),
	}
	return diskInfo, nil
}

type MountInfo struct {
	DeviceName  string   `json:"deviceName"`
	FsType      string   `json:"fsType"`
	MountPoints []string `json:"mountPoints"`
	MountParams []string `json:"mountParams"`
}

func parseMountInfo(cmdRetLine string) *MountInfo {
	splits := strings.Split(cmdRetLine, " ")
	if len(splits) != 4 {
		return nil
	}

	mountParams := strings.Split(strings.TrimRight(strings.TrimLeft(splits[3], "("), ")"), ",")
	return &MountInfo{
		DeviceName:  splits[0],
		MountPoints: []string{splits[1]},
		FsType:      splits[2],
		MountParams: mountParams,
	}
}

func parseMountInfos(cmdRet string) []*MountInfo {
	reader := strings.NewReader(cmdRet)
	scanner := bufio.NewScanner(reader)
	mountInfoMap := make(map[string]*MountInfo)
	for scanner.Scan() {
		line := scanner.Text()
		mountInfo := parseMountInfo(line)
		if _, ok := mountInfoMap[mountInfo.DeviceName]; !ok {
			mountInfoMap[mountInfo.DeviceName] = mountInfo
		} else {
			mountInfoMap[mountInfo.DeviceName].MountPoints = append(mountInfoMap[mountInfo.DeviceName].MountPoints, mountInfo.MountPoints...)
		}
	}
	mountInfos := make([]*MountInfo, 0)
	for _, info := range mountInfoMap {
		mountInfos = append(mountInfos, info)
	}

	return mountInfos
}

// getParentDevice Obtain the disk corresponding to the logical partition
func getParentDevice(partition string, allBlockDevs []string) (string, error) {
	for _, blockDev := range allBlockDevs {
		partDir := fmt.Sprintf(sysFsPartitionDir, blockDev, partition)
		isExists, err := libFile.FileExists(partDir)
		if err != nil {
			return "", err
		}
		if isExists {
			return blockDev, nil
		}
	}
	return "", errors.Errorf("%s doesn't have parent block device", partition)
}

func listAllBlockDevices() ([]string, error) {
	blockDevDirs, err := libFile.ListFiles(sysFsBaseDir, file.LinkType)
	if err != nil {
		return nil, err
	}

	blockDevs := make([]string, 0)
	for _, blockDevDir := range blockDevDirs {
		blockDevs = append(blockDevs, filepath.Base(blockDevDir))
	}
	return blockDevs, nil
}

func getRealDeviceName(devName string) (string, error) {
	absDevName, err := filepath.Abs(devName)
	if err != nil {
		return "", err
	}
	realDevName := absDevName

	evaledDevName, err := filepath.EvalSymlinks(realDevName)
	if err != nil {
		log.WithError(err).Errorf("getRealDeviceName %s EvalSymlinks failed", realDevName)
		return "", err
	}

	if evaledDevName != realDevName {
		realDevName = evaledDevName
	}

	return strings.TrimPrefix(realDevName, "/dev/"), nil
}

func getBlockDeviceName(devName string) (string, error) {
	realDevName, err := getRealDeviceName(devName)
	if err != nil {
		return "", err
	}

	devDir := fmt.Sprintf(sysFsDevDir, realDevName)
	isExists, err := libFile.FileExists(devDir)
	if err != nil {
		return "", err
	}

	if !isExists {
		allBlockDevs, err := listAllBlockDevices()
		if err != nil {
			return "", err
		}

		realDevName, err = getParentDevice(realDevName, allBlockDevs)
		if err != nil {
			return "", err
		}
	}
	return realDevName, nil
}

func getDeviceTotalSize(devName string) (uint64, error) {
	return getDeviceSize(sysFsTotalSize, devName)
}

func getDeviceSectorSize(devName string) (uint64, error) {
	return getDeviceSize(sysFsSectorSize, devName)
}

func getDeviceSize(cmdTmpl, devName string) (uint64, error) {
	getSizeCmd := fmt.Sprintf(cmdTmpl, devName)
	execResult, err := libShell.NewCommand(getSizeCmd).WithOutputType(shell.StdOutput).WithUser(shell.RootUser).Execute()
	if err != nil {
		return 0, err
	}
	sizeRaw := strings.TrimSuffix(execResult.Output, "\n")

	size, err := strconv.ParseUint(sizeRaw, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (d DiskImpl) BatchGetDiskInfos() ([]*DiskInfo, error) {
	execResult, err := libShell.NewCommand(getMountPointsCommand).WithOutputType(shell.StdOutput).WithUser(shell.RootUser).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "exec getMountPointsCommand failed")
	}
	mountInfos := parseMountInfos(execResult.Output)

	diskInfos := make([]*DiskInfo, 0)

	for _, info := range mountInfos {
		devName, err := getBlockDeviceName(info.DeviceName)
		if err != nil {
			log.WithError(err).Errorf("failed to get real device name:%s", info.DeviceName)
			continue
		}

		totalSize, err := getDeviceTotalSize(devName)
		if err != nil {
			log.WithError(err).Errorf("failed to get device total size:%s", devName)
			continue
		}

		sectorSize, err := getDeviceSectorSize(devName)
		if err != nil {
			log.WithError(err).Errorf("failed to get device sector size:%s", devName)
			continue
		}

		diskInfos = append(diskInfos, &DiskInfo{
			TotalSize:  totalSize,
			SectorSize: sectorSize,
			MountInfo:  *info,
		})
	}

	return diskInfos, nil
}
