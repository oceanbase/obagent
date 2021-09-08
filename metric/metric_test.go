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

package metric

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestMetric() Metric {
	name := "test"

	metricType := Gauge

	currentTime := time.Now()

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["k1"] = "v1"
	tags["k2"] = "v2"

	fields["f1"] = 1.0
	fields["f2"] = 2.0

	metric := NewMetric(name, fields, tags, currentTime, metricType)
	return metric
}

func TestNewMetric(t *testing.T) {
	name := "test"

	metricType := Gauge

	currentTime := time.Now()

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["k1"] = "v1"
	tags["k2"] = "v2"

	fields["f1"] = 1.0
	fields["f2"] = 2.0

	metric := NewMetric(name, fields, tags, currentTime, metricType)

	require.Equal(t, "test", metric.GetName())
	require.Equal(t, Gauge, metric.GetMetricType())
	require.Equal(t, fields, metric.Fields())
	require.Equal(t, tags, metric.Tags())
	require.Equal(t, currentTime, metric.GetTime())
}

func TestSetName(t *testing.T) {
	metric := newTestMetric()
	metric.SetName("cpu")
	require.Equal(t, "cpu", metric.GetName())
}

func TestSetTime(t *testing.T) {
	metric := newTestMetric()
	currentTime := time.Now()
	metric.SetTime(currentTime)
	require.Equal(t, currentTime, metric.GetTime())
}

func TestSetMetricType(t *testing.T) {
	metric := newTestMetric()
	metric.SetMetricType(Counter)
	require.Equal(t, Counter, metric.GetMetricType())
}

func TestCloneMetric(t *testing.T) {
	m := newTestMetric()
	mc := m.Clone()
	require.True(t, reflect.DeepEqual(m, mc))
}
