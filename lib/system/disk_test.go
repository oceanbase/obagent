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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func prepareTestDir(tree string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v\n", err)
	}

	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		return "", fmt.Errorf("error evaling temp directory: %v\n", err)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, tree), 0755)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return filepath.Join(tmpDir, tree), nil
}

func TestParseMountInfo(t *testing.T) {
	cmdRet := "/dev/v01d /data/0 ext4 (rw)"
	mountInfo := parseMountInfo(cmdRet)
	assert.Equal(t, "/dev/v01d", mountInfo.DeviceName)
	if !assert.Equal(t, 1, len(mountInfo.MountPoints)) {
		t.Fatal("check mountPoints length failed")
	}
	assert.Equal(t, "/data/0", mountInfo.MountPoints[0])
	assert.Equal(t, "ext4", mountInfo.FsType)
	if !assert.Equal(t, 1, len(mountInfo.MountParams)) {
		t.Fatal("check mountParams length failed")
	}
	assert.Equal(t, "rw", mountInfo.MountParams[0])
}

func TestParseMountInfos(t *testing.T) {
	cmdRet := `/dev/sda3 / ext4 (rw,relatime,data=ordered)
/dev/sdb /opt/k8s ext4 (rw,noatime,nodiratime,nobarrier,data=ordered)
/dev/mapper/vg_ocplog-ob7_clog /k8s/disks/clog/9d179102-f919-4683-8263-ea5272efa882 ext4 (rw,relatime,stripe=48,data=ordered)
/dev/mapper/vg_ocplog-ob6_clog /k8s/disks/clog/67d38fa7-379d-47d1-97ab-eca89a853984 ext4 (rw,relatime,stripe=48,data=ordered)
/dev/mapper/vg_ocplog-ob4_clog /k8s/disks/clog/5e9e730d-ba44-459d-9c61-b60d62f66922 ext4 (rw,relatime,stripe=48,data=ordered)
/dev/sda2 /boot ext4 (rw,relatime,data=ordered)
/dev/mapper/vg_ocpdata-ob6_data /k8s/disks/data/550ec764-67da-40c2-bf9b-239f45eba871 ext4 (rw,relatime,stripe=48,data=ordered)
/dev/mapper/vg_ocpdata-ob4_data /k8s/disks/data/cbaacc38-b8e8-451c-a9ce-90ab4348a134 ext4 (rw,relatime,stripe=48,data=ordered)
/dev/mapper/vg_ocphome-ob10_home /k8s/disks/home/0663de3c-83d8-41a9-8355-55fdad8e7876 ext4 (rw,relatime,data=ordered)
/dev/mapper/vg_ocplog-ob5_clog /k8s/disks/clog/712a0df4-5a29-48ad-8f0e-fd32c1958582 ext4 (rw,relatime,stripe=48,data=ordered)
/dev/mapper/vglog-lvlog /data/log1 ext4 (rw,noatime,nodiratime,nobarrier,stripe=32,data=ordered)`
	mountInfos := parseMountInfos(cmdRet)
	assert.Equal(t, 11, len(mountInfos))
}

func TestBatchGetDiskInfos(t *testing.T) {
	disk := DiskImpl{}
	diskInfos, err := disk.BatchGetDiskInfos()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, diskInfos)
}
