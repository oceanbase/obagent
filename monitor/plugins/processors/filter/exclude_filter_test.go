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

package filter

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

func TestMatchTags(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
        conditions:
          - metric: test
            tags:
              k1: v1
              k2: v2
    `
	var filterConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &filterConfigMap)

	excludeFilter := &ExcludeFilter{}
	excludeFilter.Init(context.Background(), filterConfigMap)
	metricsProcessed, _ := excludeFilter.Process(context.Background(), metrics...)
	require.Equal(t, 0, len(metricsProcessed))
}

func TestMatchTagsMatchAllFields(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
        conditions:
          - metric: test
            tags:
              k1: v1
              k2: v2
            anyFields:
              f1: 1.0
              f2: 3.0
    `
	var filterConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &filterConfigMap)

	excludeFilter := &ExcludeFilter{}
	excludeFilter.Init(context.Background(), filterConfigMap)
	metricsProcessed, _ := excludeFilter.Process(context.Background(), metrics...)
	require.Equal(t, 0, len(metricsProcessed))
}

func TestMatchTagsAndNotMatchAllFields(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
        conditions:
          - metric: test
            tags:
              k1: v1
              k2: v2
            fields:
              f1: 1.1
              f2: 3.1
    `
	var filterConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &filterConfigMap)

	excludeFilter := &ExcludeFilter{}
	excludeFilter.Init(context.Background(), filterConfigMap)
	metricsProcessed, _ := excludeFilter.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
}

func TestMatchTagsAndNotMatchSomeFields(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
        conditions:
          - metric: test
            tags:
              k1: v1
              k2: v2
            fields:
              f1: 1.0
              f2: 3.1
    `
	var filterConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &filterConfigMap)

	excludeFilter := &ExcludeFilter{}
	excludeFilter.Init(context.Background(), filterConfigMap)
	metricsProcessed, _ := excludeFilter.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
}

func TestTagNotMatch(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
        conditions:
          - metric: test
            tags:
              k1: v1
              k2: v3
    `
	var filterConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &filterConfigMap)

	excludeFilter := &ExcludeFilter{}
	excludeFilter.Init(context.Background(), filterConfigMap)
	metricsProcessed, _ := excludeFilter.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
}
