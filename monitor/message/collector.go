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
	"github.com/oceanbase/obagent/monitor/plugins/common"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var (
	invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_:]`)
	validNameCharRE   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)
)

type MetricFamily struct {
	Samples  []*Sample
	Type     Type
	LabelSet map[string]int
}

type Sample struct {
	Labels         map[string]string
	Value          float64
	HistogramValue map[float64]uint64
	SummaryValue   map[float64]float64
	Count          uint64
	Sum            float64
	Timestamp      time.Time
}

type CollectorConfig struct {
	ExportTimestamp    bool
	TimestampPrecision common.TimestampPrecision
}

type Collector struct {
	Fam         map[string]*MetricFamily
	ConstLabels prometheus.Labels
	config      *CollectorConfig
}

func NewCollector(config *CollectorConfig) *Collector {
	return &Collector{
		Fam:         make(map[string]*MetricFamily),
		ConstLabels: make(map[string]string),
		config:      config,
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	for name, metricFamily := range c.Fam {

		var labelNames []string
		for k := range metricFamily.LabelSet {
			labelNames = append(labelNames, k)
		}
		desc := prometheus.NewDesc(name, "monitor collected message", labelNames, c.ConstLabels)
		for _, sample := range metricFamily.Samples {
			m, err := CreateMetric(labelNames, c.config, sample, metricFamily, desc)
			if err != nil {
				log.WithError(err).Warn("prometheus create message failed")
				continue
			}
			ch <- m
		}
	}
}

func CreateMetric(labelNames []string, config *CollectorConfig, sample *Sample, metricFamily *MetricFamily, desc *prometheus.Desc) (prometheus.Metric, error) {
	var labels []string
	for _, label := range labelNames {
		v := sample.Labels[label]
		labels = append(labels, v)
	}
	var m prometheus.Metric
	var err error
	switch metricFamily.Type {
	case Summary:
		m, err = prometheus.NewConstSummary(desc, sample.Count, sample.Sum, sample.SummaryValue, labels...)
	case Histogram:
		m, err = prometheus.NewConstHistogram(desc, sample.Count, sample.Sum, sample.HistogramValue, labels...)
	default:
		m, err = prometheus.NewConstMetric(desc, getPromValueType(metricFamily.Type), sample.Value, labels...)
	}
	if config != nil && config.ExportTimestamp && config.TimestampPrecision == common.Millisecond {
		m = millisecondTimestampedMetric{Metric: m, t: sample.Timestamp}
	} else if config != nil && config.ExportTimestamp && config.TimestampPrecision == common.Second {
		m = secondTimestampedMetric{Metric: m, t: sample.Timestamp}
	}
	return m, err
}

func CreateMetricFamily(metrics []*Message) map[string]*MetricFamily {
	mfs := make(map[string]*MetricFamily)
	uniqueMetrics := UniqueMetrics(metrics)
	for _, m := range uniqueMetrics {
		labels := make(map[string]string)
		for _, e := range m.Tags() {
			labels[e.Name] = e.Value
		}
		switch m.GetMetricType() {
		case Summary:
			var name string
			var sum float64
			var count uint64
			summaryValue := make(map[float64]float64)
			for _, fe := range m.Fields() {
				var value float64
				switch fv := fe.Value.(type) {
				case int64:
					value = float64(fv)
				case uint64:
					value = float64(fv)
				case float64:
					value = fv
				default:
					continue
				}

				switch fe.Name {
				case "sum":
					sum = value
				case "count":
					count = uint64(value)
				default:
					limit, err := strconv.ParseFloat(fe.Name, 64)
					if err == nil {
						summaryValue[limit] = value
					}
				}
			}
			sample := &Sample{
				Labels:       labels,
				SummaryValue: summaryValue,
				Count:        count,
				Sum:          sum,
				Timestamp:    m.GetTime(),
			}
			name = sanitize(m.GetName())

			if !isValidTagName(name) {
				continue
			}

			addMetricFamily(mfs, m, sample, name)

		case Histogram:
			var name string
			var sum float64
			var count uint64
			histogramValue := make(map[float64]uint64)
			for _, fe := range m.Fields() {
				var value float64
				switch fv := fe.Value.(type) {
				case int64:
					value = float64(fv)
				case uint64:
					value = float64(fv)
				case float64:
					value = fv
				default:
					continue
				}

				switch fe.Name {
				case "sum":
					sum = value
				case "count":
					count = uint64(value)
				default:
					limit, err := strconv.ParseFloat(fe.Name, 64)
					if err == nil {
						histogramValue[limit] = uint64(value)
					}
				}
			}
			sample := &Sample{
				Labels:         labels,
				HistogramValue: histogramValue,
				Count:          count,
				Sum:            sum,
				Timestamp:      m.GetTime(),
			}
			name = sanitize(m.GetName())

			if !isValidTagName(name) {
				continue
			}

			addMetricFamily(mfs, m, sample, name)

		default:
			for _, fe := range m.Fields() {
				var value float64
				switch fv := fe.Value.(type) {
				case int64:
					value = float64(fv)
				case uint64:
					value = float64(fv)
				case float64:
					value = fv
				default:
					continue
				}

				sample := &Sample{
					Labels:    labels,
					Value:     value,
					Timestamp: m.GetTime(),
				}
				var name string
				switch m.GetMetricType() {
				case Counter:
					if fe.Name == "counter" {
						name = sanitize(m.GetName())
					}
				case Gauge:
					if fe.Name == "gauge" {
						name = sanitize(m.GetName())
					}
				}
				if name == "" {
					if fe.Name == "value" {
						name = sanitize(m.GetName())
					} else {
						name = sanitize(fmt.Sprintf("%s_%s", m.GetName(), fe.Name))
					}
				}
				if !isValidTagName(name) {
					continue
				}

				addMetricFamily(mfs, m, sample, name)

			}
		}
	}
	return mfs
}

func addMetricFamily(mfs map[string]*MetricFamily, m *Message, sample *Sample, name string) {
	var metricFamily *MetricFamily
	var exist bool
	metricFamily, exist = mfs[name]
	if !exist {
		metricFamily = &MetricFamily{
			Samples:  make([]*Sample, 0),
			Type:     m.GetMetricType(),
			LabelSet: make(map[string]int),
		}
		mfs[name] = metricFamily
	}
	addSample(metricFamily, sample)
}

func ProcessFields(msgs []*Message) []*Message {
	var newMsgs = make([]*Message, 0)
	uniqueMetrics := UniqueMetrics(msgs)
	for _, msg := range uniqueMetrics {
		for _, field := range msg.Fields() {
			var value float64
			switch fv := field.Value.(type) {
			case int64:
				value = float64(fv)
			case uint64:
				value = float64(fv)
			case float64:
				value = fv
			default:
				continue
			}
			var name string
			switch msg.GetMetricType() {
			case Counter:
				if field.Name == "counter" {
					name = sanitize(msg.GetName())
				}
			case Gauge:
				if field.Name == "gauge" {
					name = sanitize(msg.GetName())
				}
			}
			if name == "" {
				if field.Name == "value" {
					name = sanitize(msg.GetName())
				} else {
					name = sanitize(fmt.Sprintf("%s_%s", msg.GetName(), field.Name))
				}
			}
			if !isValidTagName(name) {
				continue
			}
			var entry = FieldEntry{"value", value}
			tmpMsg := NewMessageWithTagsFields(name, msg.GetMetricType(), time.Now(), msg.Tags(), []FieldEntry{entry})
			newMsgs = append(newMsgs, tmpMsg)
		}
	}
	return newMsgs
}

func addSample(fam *MetricFamily, sample *Sample) {
	for k := range sample.Labels {
		fam.LabelSet[k]++
	}

	fam.Samples = append(fam.Samples, sample)
}

func sanitize(value string) string {
	return invalidNameCharRE.ReplaceAllString(value, "_")
}

func isValidTagName(tag string) bool {
	return validNameCharRE.MatchString(tag)
}

func getPromValueType(t Type) prometheus.ValueType {
	switch t {
	case Counter:
		return prometheus.CounterValue
	case Gauge:
		return prometheus.GaugeValue
	default:
		return prometheus.UntypedValue
	}
}

type millisecondTimestampedMetric struct {
	prometheus.Metric
	t time.Time
}

func (m millisecondTimestampedMetric) Write(pb *dto.Metric) error {
	e := m.Metric.Write(pb)
	pb.TimestampMs = proto.Int64(m.t.UnixNano() / int64(time.Millisecond))
	return e
}

type secondTimestampedMetric struct {
	prometheus.Metric
	t time.Time
}

func (m secondTimestampedMetric) Write(pb *dto.Metric) error {
	e := m.Metric.Write(pb)
	pb.TimestampMs = proto.Int64(m.t.UnixNano() / int64(time.Second))
	return e
}
