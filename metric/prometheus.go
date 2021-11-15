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
	"fmt"
	"math"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func ParseFromMetricFamily(metricFamily *dto.MetricFamily) []Metric {
	var metrics []Metric
	now := time.Now()
	for _, m := range metricFamily.Metric {
		tags := makeLabels(m, nil)
		var fields map[string]interface{}

		switch metricFamily.GetType() {

		case dto.MetricType_SUMMARY:
			fields = makeQuantiles(m)
			fields["count"] = float64(m.GetSummary().GetSampleCount())
			fields["sum"] = m.GetSummary().GetSampleSum()

		case dto.MetricType_HISTOGRAM:
			fields = makeBuckets(m)
			fields["count"] = float64(m.GetHistogram().GetSampleCount())
			fields["sum"] = m.GetHistogram().GetSampleSum()

		default:
			fields = getNameAndValue(m)
		}

		if len(fields) > 0 {
			var t time.Time
			if m.TimestampMs != nil && *m.TimestampMs > 0 {
				t = time.Unix(0, *m.TimestampMs*1000000)
			} else {
				t = now
			}
			newMetric := NewMetric(metricFamily.GetName(), fields, tags, t, valueType(metricFamily.GetType()))
			metrics = append(metrics, newMetric)
		}
	}
	return metrics
}

func makeLabels(m *dto.Metric, defaultTags map[string]string) map[string]string {
	result := map[string]string{}

	for key, value := range defaultTags {
		result[key] = value
	}

	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}

	return result
}

func makeQuantiles(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			fields[fmt.Sprint(q.GetQuantile())] = q.GetValue()
		}
	}
	return fields
}

func makeBuckets(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	for _, b := range m.GetHistogram().Bucket {
		fields[fmt.Sprint(b.GetUpperBound())] = float64(b.GetCumulativeCount())
	}
	return fields
}

func getNameAndValue(m *dto.Metric) map[string]interface{} {
	fields := make(map[string]interface{})
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields["gauge"] = m.GetGauge().GetValue()
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			fields["counter"] = m.GetCounter().GetValue()
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			fields["value"] = m.GetUntyped().GetValue()
		}
	}
	return fields
}

func valueType(metricType dto.MetricType) Type {
	switch metricType {
	case dto.MetricType_COUNTER:
		return Counter
	case dto.MetricType_GAUGE:
		return Gauge
	case dto.MetricType_SUMMARY:
		return Summary
	case dto.MetricType_HISTOGRAM:
		return Histogram
	default:
		return Untyped
	}
}
