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

package engine

import (
	"bytes"
	"time"

	"github.com/oceanbase/obagent/metric"
	"github.com/oceanbase/obagent/plugins"
)

const (
	Test metric.Type = "Test"
)

type testInput struct {
}

func (t *testInput) Init(_ map[string]interface{}) error {
	return nil
}

func (t *testInput) Close() error {
	return nil
}

func (t *testInput) SampleConfig() string {
	return "this is test input"
}

func (t *testInput) Description() string {
	return "this is test input"
}

func (t *testInput) Collect() ([]metric.Metric, error) {
	return []metric.Metric{metric.NewMetric("test", map[string]interface{}{"test": "test"}, map[string]string{"test": "test"}, time.Time{}, Test)}, nil
}

type testProcessor struct {
}

func (t *testProcessor) Init(_ map[string]interface{}) error {
	return nil
}

func (t *testProcessor) Close() error {
	return nil
}

func (t *testProcessor) SampleConfig() string {
	return "this is test processor"
}

func (t *testProcessor) Description() string {
	return "this is test processor"
}

func (t *testProcessor) Process(metrics ...metric.Metric) ([]metric.Metric, error) {
	return metrics, nil
}

type testOutput struct {
}

func (t *testOutput) Init(_ map[string]interface{}) error {
	return nil
}

func (t *testOutput) Close() error {
	return nil
}

func (t *testOutput) SampleConfig() string {
	return "this is test output"
}

func (t *testOutput) Description() string {
	return "this is test output"
}

func (t *testOutput) Write(_ []metric.Metric) error {
	return nil
}

type testExporter struct {
}

func (t *testExporter) Init(_ map[string]interface{}) error {
	return nil
}

func (t *testExporter) Close() error {
	return nil
}

func (t *testExporter) SampleConfig() string {
	return "this is test exporter"
}

func (t *testExporter) Description() string {
	return "this is test exporter"
}

func (t *testExporter) Export(_ []metric.Metric) (*bytes.Buffer, error) {
	return bytes.NewBuffer([]byte("this is test exporter")), nil
}

func init() {
	plugins.GetInputManager().Register("test", func() plugins.Input {
		return &testInput{}
	})
	plugins.GetProcessorManager().Register("test", func() plugins.Processor {
		return &testProcessor{}
	})
	plugins.GetOutputManager().Register("test", func() plugins.Output {
		return &testOutput{}
	})
	plugins.GetExporterManager().Register("test", func() plugins.Exporter {
		return &testExporter{}
	})
}

var testJSONModule = `
{
	"module": "test",
	"testInput": {
		"plugin": "test",
		"config": {
			"timeout": 10,
			"pluginConfig": null
		}
	},
	"testProcessor": {
		"plugin": "test",
		"config": {
			"timeout": 10,
			"pluginConfig": null
		}
	},
	"testOutput": {
		"plugin": "test",
		"config": {
			"timeout": 10,
			"pluginConfig": null
		}
	},
	"testExporter": {
		"plugin": "test",
		"config": {
			"timeout": 10,
			"pluginConfig": null
		}
	},
	"pipelines": [{
			"name": "pipeline1",
			"config": {
				"scheduleStrategy": "trigger",
				"exposeUrl": "/metrics/test"
			},
			"structure": {
				"inputs": [{
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}],
				"processors": [{
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}],
				"exporter": {
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}
			}
		},
		{
			"name": "pipeline2",
			"config": {
				"scheduleStrategy": "periodic",
				"period": 5
			},
			"structure": {
				"inputs": [{
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}],
				"processors": [{
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}],
				"output": {
					"plugin": "test",
					"config": {
						"timeout": 10,
						"pluginConfig": null
					}
				}
			}
		}
	]
}
`
