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
	"testing"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/metric"
)

func TestInstance(_ *testing.T) {
	testPluginConfig := &config.PluginConfig{Timeout: 0}
	testInput := &InputInstance{
		PluginName: "test",
		Config:     testPluginConfig,
	}
	testProcessor := &ProcessorInstance{
		PluginName: "test",
		Config:     testPluginConfig,
	}
	testOutput := &OutputInstance{
		PluginName: "test",
		Config:     testPluginConfig,
	}
	testExporter := &ExporterInstance{
		PluginName: "test",
		Config:     testPluginConfig,
	}

	_ = testInput.init(testPluginConfig)
	_ = testProcessor.init(testPluginConfig)
	_ = testOutput.init(testPluginConfig)
	_ = testExporter.init(testPluginConfig)

	_, _ = testInput.Collect()
	_, _ = testProcessor.Process([]metric.Metric{}...)
	_ = testOutput.Write([]metric.Metric{})
	_, _ = testExporter.Export([]metric.Metric{})

	testPluginConfig = &config.PluginConfig{Timeout: 1}
	testInput = &InputInstance{
		PluginName: "test",
		Config:     testPluginConfig,
	}
	testProcessor = &ProcessorInstance{
		PluginName: "test",
		Config:     testPluginConfig,
	}
	testOutput = &OutputInstance{
		PluginName: "test",
		Config:     testPluginConfig,
	}
	testExporter = &ExporterInstance{
		PluginName: "test",
		Config:     testPluginConfig,
	}

	_ = testInput.init(testPluginConfig)
	_ = testProcessor.init(testPluginConfig)
	_ = testOutput.init(testPluginConfig)
	_ = testExporter.init(testPluginConfig)

	_, _ = testInput.Collect()
	_, _ = testProcessor.Process([]metric.Metric{}...)
	_ = testOutput.Write([]metric.Metric{})
	_, _ = testExporter.Export([]metric.Metric{})
}
