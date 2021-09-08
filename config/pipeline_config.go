// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package config

import (
	"time"
)

type ScheduleStrategy string
type PipelineModuleStatus string

const (
	Trigger  ScheduleStrategy = "trigger"
	Periodic ScheduleStrategy = "periodic"
)

const (
	ACTIVE   PipelineModuleStatus = "active"
	INACTIVE PipelineModuleStatus = "inactive"
)

type PluginConfig struct {
	Timeout           time.Duration          `yaml:"timeout"`
	PluginInnerConfig map[string]interface{} `yaml:"pluginConfig"`
}

type PluginNode struct {
	Plugin string        `yaml:"plugin"`
	Config *PluginConfig `yaml:"config"`
}

type PipelineConfig struct {
	ScheduleStrategy ScheduleStrategy `yaml:"scheduleStrategy"`
	ExposeUrl        string           `yaml:"exposeUrl"`
	Period           time.Duration    `yaml:"period"`
	DownSamplePeriod time.Duration    `yaml:"downSamplePeriod"`
}

type PipelineStructure struct {
	Inputs     []*PluginNode `yaml:"inputs"`
	Processors []*PluginNode `yaml:"processors"`
	Output     *PluginNode   `yaml:"output"`
	Exporter   *PluginNode   `yaml:"exporter"`
}

type PipelineNode struct {
	Name      string             `yaml:"name"`
	Config    *PipelineConfig    `yaml:"config"`
	Structure *PipelineStructure `yaml:"structure"`
}

type PipelineModule struct {
	Name      string               `yaml:"name"`
	Status    PipelineModuleStatus `yaml:"status"`
	Pipelines []*PipelineNode      `yaml:"pipelines"`
}
