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

package agentctl

import "github.com/oceanbase/obagent/config"

// agentctl meta
type AgentctlConfig struct {
	SDKConfig config.SDKConfig `yaml:"sdkConfig"`
	// log config
	Log config.LogConfig `yaml:"log"`

	RunDir       string `yaml:"runDir"`
	ConfDir      string `yaml:"confDir"`
	LogDir       string `yaml:"logDir"`
	BackupDir    string `yaml:"backupDir"`
	TempDir      string `yaml:"tempDir"`
	TaskStoreDir string `yaml:"taskStoreDir"`
	AgentPkgName string `yaml:"agentPkgName"`
	PkgExt       string `yaml:"pkgExt"`
	PkgStoreDir  string `yaml:"pkgStoreDir"`
}
