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

package message

import (
	"fmt"
	"math"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func ParseFromMetricFamily(metricFamily *dto.MetricFamily) []*Message {
	var metrics []*Message
	now := time.Now()
	for _, m := range metricFamily.Metric {
		tags := makeLabels(m)
		var fields []FieldEntry

		switch metricFamily.GetType() {

		case dto.MetricType_SUMMARY:
			fields = makeQuantiles(m)
			fields = append(fields, FieldEntry{Name: "count", Value: float64(m.GetSummary().GetSampleCount())})
			fields = append(fields, FieldEntry{Name: "sum", Value: m.GetSummary().GetSampleSum()})
		case dto.MetricType_HISTOGRAM:
			fields = makeBuckets(m)
			fields = append(fields, FieldEntry{Name: "count", Value: float64(m.GetHistogram().GetSampleCount())})
			fields = append(fields, FieldEntry{Name: "sum", Value: m.GetHistogram().GetSampleSum()})
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
			newMetric := NewMessageWithTagsFields(metricFamily.GetName(), ValueType(metricFamily.GetType()), t, tags, fields)
			metrics = append(metrics, newMetric)
		}
	}
	return metrics
}

func makeLabels(m *dto.Metric) []TagEntry {
	result := make([]TagEntry, 0, len(m.Label))
	for _, lp := range m.Label {
		result = append(result, TagEntry{Name: lp.GetName(), Value: lp.GetValue()})
	}
	return result
}

func makeQuantiles(m *dto.Metric) []FieldEntry {
	fields := make([]FieldEntry, 0, len(m.GetSummary().Quantile)+2)
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			fields = append(fields, FieldEntry{Name: fmt.Sprint(q.GetQuantile()), Value: q.GetValue()})
		}
	}
	return fields
}

func makeBuckets(m *dto.Metric) []FieldEntry {
	fields := make([]FieldEntry, 0, len(m.GetHistogram().Bucket)+2)
	for _, b := range m.GetHistogram().Bucket {
		fields = append(fields, FieldEntry{Name: fmt.Sprint(b.GetUpperBound()), Value: float64(b.GetCumulativeCount())})
	}
	return fields
}

func getNameAndValue(m *dto.Metric) []FieldEntry {
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			return []FieldEntry{{Name: "gauge", Value: m.GetGauge().GetValue()}}
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			return []FieldEntry{{Name: "counter", Value: m.GetCounter().GetValue()}}
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			return []FieldEntry{{Name: "value", Value: m.GetUntyped().GetValue()}}
		}
	}
	return []FieldEntry{}
}

func ValueType(metricType dto.MetricType) Type {
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
