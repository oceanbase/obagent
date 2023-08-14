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

package label

import (
	"context"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/shell"
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/plugins/common"
)

const sampleConfig = `
    labelTags:
      installPath: /home/admin/oceanbase
      dataDiskPath: /data/1
      logDiskPath: /data/log1
`

const description = `
    check mountpoint with labelTags, add mount_label/is_ob_disk if true
`

var (
	installMountPath, dataMountPath, logMountPath string
	checkReadonlyDisks                            map[string]bool

	mountPV map[string]string
	env     string
)

type MountLabelConfig struct {
	LabelTags map[string]string `yaml:"labelTags"`
	ObStatus  string            `yaml:"ob_status"`
}

type MountLabelProcessor struct {
	Config *MountLabelConfig
}

func (m *MountLabelProcessor) SampleConfig() string {
	return sampleConfig
}

func (m *MountLabelProcessor) Description() string {
	return description
}

func (m *MountLabelProcessor) Init(ctx context.Context, config map[string]interface{}) error {
	var mountLabelConfig MountLabelConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "mountLabelProcessor encode config")
	}
	err = yaml.Unmarshal(configBytes, &mountLabelConfig)
	if err != nil {
		return errors.Wrap(err, "mountLabelProcessor decode config")
	}

	m.Config = &mountLabelConfig
	log.WithContext(ctx).Infof("init mountLabelProcessor with config: %+v", m.Config)

	installPath, err := filepath.EvalSymlinks(m.Config.LabelTags["installPath"])
	if err != nil {
		log.WithContext(ctx).Warn("check installPath failed")
	}
	dataDiskPath, err := filepath.EvalSymlinks(m.Config.LabelTags["dataDiskPath"])
	if err != nil {
		log.WithContext(ctx).Warn("check dataDiskPath failed")
	}
	logDiskPath, err := filepath.EvalSymlinks(m.Config.LabelTags["logDiskPath"])
	if err != nil {
		log.WithContext(ctx).Warn("check logDiskPath failed")
	}
	installMountPath = common.GetMountPath(installPath)
	dataMountPath = common.GetMountPath(dataDiskPath)
	logMountPath = common.GetMountPath(logDiskPath)

	checkReadonly := m.Config.LabelTags["checkReadonly"]
	checkReadonlyDisks = parseDisk(checkReadonly)

	env, err = common.CheckNodeEnv(ctx)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("check node env failed")
	}
	mountPV, err = getMountPV(ctx)
	if err != nil {
		log.WithContext(ctx).Warn("get mountPV failed")
	}
	return nil
}

func (m *MountLabelProcessor) Start(in <-chan []*message.Message, out chan<- []*message.Message) error {
	for msgs := range in {
		newMsgs, err := m.Process(context.Background(), msgs...)
		if err != nil {
			log.Errorf("mountLabel process messages failed, err: %s", err)
		}
		out <- newMsgs
	}
	return nil
}

func (m *MountLabelProcessor) Stop() {}

func (m *MountLabelProcessor) Process(ctx context.Context, metrics ...*message.Message) ([]*message.Message, error) {
	for _, metric := range metrics {
		v, mountpointFound := metric.GetTag("mountpoint")
		if mountpointFound {
			if v == installMountPath {
				metric.AddTag("mount_label", "install_path")
				metric.AddTag("install_path", m.Config.LabelTags["installPath"])
			} else if v == dataMountPath {
				metric.AddTag("mount_label", "data_disk_path")
				metric.AddTag("data_disk_path", m.Config.LabelTags["dataDiskPath"])
			} else if v == logMountPath {
				metric.AddTag("mount_label", "log_disk_path")
				metric.AddTag("log_disk_path", m.Config.LabelTags["logDiskPath"])
			}
			if checkReadonlyDisks[v] {
				metric.AddTag("check_readonly", "1")
			}
			if (v == installMountPath || v == dataMountPath || v == logMountPath) && m.Config.ObStatus == "active" {
				metric.AddTag("is_ob_disk", "1")
			} else {
				metric.AddTag("is_ob_disk", "0")
			}
		}

		if env == common.Container || m.Config.ObStatus == "inactive" {
			continue
		}
		var device string
		var obDevices = make([]string, 0)
		if dev, devFound := metric.GetTag("dev"); devFound {
			device = dev
		} else {
			dev, found := metric.GetTag("device")
			if found {
				device = dev
			}
		}
		if device != "" {
			dataPV, ok := mountPV[dataMountPath]
			if ok {
				if strings.Contains(dataPV, device) {
					obDevices = append(obDevices, "ob_data")
				}
			} else {
				log.Warnf("can not found PV of dataMountPath %s", dataMountPath)
			}
			logPV, ok := mountPV[logMountPath]
			if ok {
				if strings.Contains(logPV, device) {
					obDevices = append(obDevices, "ob_log")
				}
			} else {
				log.Warnf("can not found PV of logMountPath %s", logMountPath)
			}
			if len(obDevices) > 0 {
				obDevice := strings.Join(obDevices, ",")
				metric.AddTag("ob_device", obDevice)
				metric.AddTag("is_ob_volume", "1")
			}
		}
	}

	return metrics, nil
}

func getMountPV(ctx context.Context) (map[string]string, error) {
	var mountPV = make(map[string]string)
	cmd := "df -h"
	command := shell.ShellImpl{}.NewCommand(cmd)
	result, err := command.ExecuteWithDebug()
	if err != nil {
		log.WithContext(ctx).Warnf("execute cmd %s failed, err: %s", cmd, err)
		return nil, err
	}
	lines := strings.Split(result.Output, "\n")
	for i := 1; i < len(lines); i++ {
		if len(lines[i]) == 0 {
			continue
		}
		headers := strings.Fields(lines[i])
		fileSystem := headers[0]
		mount := headers[len(headers)-1]
		pv, err := filepath.EvalSymlinks(fileSystem)
		if err != nil {
			log.WithContext(ctx).Warnf("get pv of fileSystem %s failed, err: %s", fileSystem, err)
			continue
		}
		if _, ok := mountPV[mount]; ok {
			log.WithContext(ctx).Warnf("mount %s is repeat", mount)
			continue
		}
		mountPV[mount] = pv
	}
	return mountPV, nil
}

func parseDisk(disk string) map[string]bool {
	disks := strings.Split(disk, "|")
	mountDisks := make(map[string]bool, len(disks))
	for _, it := range disks {
		mountDisks[common.GetMountPath(strings.Trim(it, " "))] = true
	}
	return mountDisks
}
