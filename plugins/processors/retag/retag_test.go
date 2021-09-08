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

package retag

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/metric"
)

func newTestMetric() metric.Metric {
	name := "test"

	metricType := metric.Gauge

	currentTime := time.Now()

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["k1"] = "v1"
	tags["k2"] = "v2"

	fields["f1"] = 1.0
	fields["f2"] = 2.0

	return metric.NewMetric(name, fields, tags, currentTime, metricType)
}

func newTestMetrics() []metric.Metric {
	metricEntry := newTestMetric()
	var metrics []metric.Metric
	metrics = append(metrics, metricEntry)
	return metrics
}

func TestAddNewTag(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      newTags:
        k3: "v3"
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 3, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.Tags()["k3"]
	require.Equal(t, "v3", v)
	require.True(t, found)
}

func TestRenameTag(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      renameTags:
        k1: k3
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.Tags()["k1"]
	require.True(t, !found)
	v, found = metricEntryProcessed.Tags()["k3"]
	require.True(t, found)
	require.Equal(t, "v1", v)
}

func TestCopyTag(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      copyTags:
        k1: k3
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 3, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.Tags()["k1"]
	require.True(t, found)
	require.Equal(t, "v1", v)
	v, found = metricEntryProcessed.Tags()["k3"]
	require.True(t, found)
	require.Equal(t, "v1", v)
}

func TestAddNewTagAlreadyExists(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      newTags:
        k1: "v3"
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.Tags()["k1"]
	require.Equal(t, "v1", v)
	require.True(t, found)
}

func TestRenameTagAlreadyExists(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      renameTags:
        k1: "k2"
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.Tags()["k1"]
	require.Equal(t, "v1", v)
	require.True(t, found)
	v, found = metricEntryProcessed.Tags()["k2"]
	require.Equal(t, "v2", v)
	require.True(t, found)
}

func TestCopyTagAlreadyExists(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      copyTags:
        k1: "k2"
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.Tags()["k1"]
	require.Equal(t, "v1", v)
	require.True(t, found)
	v, found = metricEntryProcessed.Tags()["k2"]
	require.Equal(t, "v2", v)
	require.True(t, found)
}

func TestRenameTagNotExists(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      renameTags:
        k3: "k2"
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.Tags()["k1"]
	require.Equal(t, "v1", v)
	require.True(t, found)
	v, found = metricEntryProcessed.Tags()["k2"]
	require.Equal(t, "v2", v)
	require.True(t, found)
}

func TestCopyTagNotExists(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      copyTags:
        k3: "k2"
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.Tags()["k1"]
	require.Equal(t, "v1", v)
	require.True(t, found)
	v, found = metricEntryProcessed.Tags()["k2"]
	require.Equal(t, "v2", v)
	require.True(t, found)
}
