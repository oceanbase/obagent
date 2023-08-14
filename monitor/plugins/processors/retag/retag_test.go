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

package retag

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/monitor/message"
)

func newTestMetric() *message.Message {
	name := "test"

	metricType := message.Gauge

	currentTime := time.Now()

	return message.NewMessage(name, metricType, currentTime).
		AddTag("k1", "v1").AddTag("k2", "v2").
		AddField("f1", 1.0).AddField("f2", 2.0)
}

func newTestMetrics() []*message.Message {
	metricEntry := newTestMetric()
	var metrics []*message.Message
	metrics = append(metrics, metricEntry)
	return metrics
}

func TestAddNewTag(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      newTags:
        k3: "v3"
        empty-value:
    `
	var retagConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &retagConfigMap)

	retagProcessor := &RetagProcessor{}
	retagProcessor.Init(context.Background(), retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 4, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.GetTag("k3")
	require.Equal(t, "v3", v)
	require.True(t, found)
	emptyValue, emptyValueFound := metricEntryProcessed.GetTag("empty-value")
	require.Equal(t, "", emptyValue)
	require.True(t, emptyValueFound)
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
	retagProcessor.Init(context.Background(), retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.GetTag("k1")
	require.True(t, !found)
	v, found = metricEntryProcessed.GetTag("k3")
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
	retagProcessor.Init(context.Background(), retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 3, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.GetTag("k1")
	require.True(t, found)
	require.Equal(t, "v1", v)
	v, found = metricEntryProcessed.GetTag("k3")
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
	retagProcessor.Init(context.Background(), retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.GetTag("k1")
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
	retagProcessor.Init(context.Background(), retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.GetTag("k1")
	require.Equal(t, "v1", v)
	require.True(t, found)
	v, found = metricEntryProcessed.GetTag("k2")
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
	retagProcessor.Init(context.Background(), retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.GetTag("k1")
	require.Equal(t, "v1", v)
	require.True(t, found)
	v, found = metricEntryProcessed.GetTag("k2")
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
	retagProcessor.Init(context.Background(), retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.GetTag("k1")
	require.Equal(t, "v1", v)
	require.True(t, found)
	v, found = metricEntryProcessed.GetTag("k2")
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
	retagProcessor.Init(context.Background(), retagConfigMap)
	metricsProcessed, _ := retagProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	metricEntryProcessed := metricsProcessed[0]
	require.Equal(t, 2, len(metricEntryProcessed.Tags()))
	v, found := metricEntryProcessed.GetTag("k1")
	require.Equal(t, "v1", v)
	require.True(t, found)
	v, found = metricEntryProcessed.GetTag("k2")
	require.Equal(t, "v2", v)
	require.True(t, found)
}
