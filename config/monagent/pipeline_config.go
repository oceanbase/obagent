package monagent

import (
	"time"
)

type ScheduleStrategy string
type PipelineModuleStatus string

const (
	BySource ScheduleStrategy = "bySource"
)

const (
	ACTIVE   PipelineModuleStatus = "active"
	INACTIVE PipelineModuleStatus = "inactive"
)

func (p PipelineModuleStatus) Validate() bool {
	return p == ACTIVE || p == INACTIVE
}

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
