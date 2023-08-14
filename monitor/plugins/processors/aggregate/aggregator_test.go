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

package aggregate

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
	msg := message.NewMessage(name, metricType, time.Now()).
		AddTag("k1", "v1").
		AddTag("k2", "v2").
		AddField("f1", 1.0).
		AddField("f2", 2.0)
	return msg
}

func newTestMetrics(count int) []*message.Message {
	var metrics []*message.Message
	for i := 0; i < count; i++ {
		metricEntry := newTestMetric()
		metrics = append(metrics, metricEntry)
	}
	return metrics
}

func TestDuplicate(t *testing.T) {
	metrics := newTestMetrics(2)
	configStr := `
        rules:
          - metric: test
            tags: [ k1 ]
    `
	var aggregatorConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &aggregatorConfigMap)

	aggregator := &Aggregator{}
	aggregator.Init(context.Background(), aggregatorConfigMap)

	metricsProcessed, _ := aggregator.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	v, exists := metricsProcessed[0].GetField("f1")
	require.True(t, exists)
	f, ok := v.(float64)
	require.True(t, ok)
	require.Equal(t, 1.0, f)
}

func TestMatch(t *testing.T) {
	metrics := newTestMetrics(2)
	metrics[1].SetTag("k2", "vvvvv")
	configStr := `
        rules:
          - metric: test
            tags: [ k1 ]
    `
	var aggregatorConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &aggregatorConfigMap)

	aggregator := &Aggregator{}
	aggregator.Init(context.Background(), aggregatorConfigMap)

	metricsProcessed, _ := aggregator.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
	v, exists := metricsProcessed[0].GetField("f1")
	require.True(t, exists)
	f, ok := v.(float64)
	require.True(t, ok)
	require.Equal(t, 2.0, f)
}

func TestNotMatch(t *testing.T) {
	t.Skip()
	metrics := newTestMetrics(2)
	metrics[1].SetTag("k2", "vvvvv")
	configStr := `
        rules:
          - metric: test
            tags: [ t1 ]
    `
	var aggregatorConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &aggregatorConfigMap)

	aggregator := &Aggregator{}
	aggregator.Init(context.Background(), aggregatorConfigMap)

	metricsProcessed, _ := aggregator.Process(context.Background(), metrics...)
	require.Equal(t, 2, len(metricsProcessed))
	v, exists := metricsProcessed[0].GetField("f1")
	require.True(t, exists)
	f, ok := v.(float64)
	require.True(t, ok)
	require.Equal(t, 1.0, f)
}
