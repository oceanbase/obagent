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

package prometheus

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/metric"
)

const sampleConfig = `
formatType: fmtText
`

const description = `
export data using prometheus protocol
`

var (
	invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_:]`)
	validNameCharRE   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)
)

type Config struct {
	FormatType string `yaml:"formatType"`
}

type Prometheus struct {
	sourceConfig map[string]interface{}

	config *Config
}

type MetricFamily struct {
	Samples  []*Sample
	Type     metric.Type
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

type Collector struct {
	fam map[string]*MetricFamily
}

var Format = map[string]expfmt.Format{
	"fmtText": expfmt.FmtText,
}

func (p *Prometheus) Init(config map[string]interface{}) error {
	p.sourceConfig = config
	configData, err := yaml.Marshal(p.sourceConfig)
	if err != nil {
		return errors.Wrap(err, "prometheus exporter encode config")
	}
	p.config = &Config{}
	err = yaml.Unmarshal(configData, p.config)
	if err != nil {
		return errors.Wrap(err, "prometheus exporter decode config")
	}
	log.Infof("prometheus exporter config : %v", p.config)
	_, exist := Format[p.config.FormatType]
	if !exist {
		return errors.New("format type not exist")
	}
	return nil
}

func (p *Prometheus) Close() error {
	return nil
}

func (p *Prometheus) SampleConfig() string {
	return sampleConfig
}

func (p *Prometheus) Description() string {
	return description
}

func (p *Prometheus) Export(metrics []metric.Metric) (*bytes.Buffer, error) {
	collector := NewCollector()
	collector.fam = createMetricFamily(metrics)
	registry := prometheus.NewRegistry()
	err := registry.Register(collector)
	if err != nil {
		return nil, errors.Wrap(err, "exporter prometheus register collector")
	}
	metricFamilies, err := registry.Gather()
	if err != nil {
		return nil, errors.Wrap(err, "exporter prometheus registry gather")
	}
	buffer := bytes.NewBuffer(make([]byte, 0, 4096))
	encoder := expfmt.NewEncoder(buffer, Format[p.config.FormatType])
	for _, metricFamily := range metricFamilies {
		err := encoder.Encode(metricFamily)
		if err != nil {
			log.WithError(err).Error("exporter encode metricFamily failed")
			continue
		}
	}
	return buffer, nil
}

func NewCollector() *Collector {
	return &Collector{
		fam: make(map[string]*MetricFamily),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	for name, metricFamily := range c.fam {

		var labelNames []string
		for k := range metricFamily.LabelSet {
			labelNames = append(labelNames, k)
		}
		desc := prometheus.NewDesc(name, "monitor collected metric", labelNames, nil)
		for _, sample := range metricFamily.Samples {
			m, err := createMetric(labelNames, sample, metricFamily, desc)
			if err != nil {
				log.WithError(err).Error("prometheus create metric failed")
				continue
			}
			ch <- m
		}
	}
}

func createMetric(labelNames []string, sample *Sample, metricFamily *MetricFamily, desc *prometheus.Desc) (prometheus.Metric, error) {
	var labels []string
	for _, label := range labelNames {
		v := sample.Labels[label]
		labels = append(labels, v)
	}
	var m prometheus.Metric
	var err error
	switch metricFamily.Type {
	case metric.Summary:
		m, err = prometheus.NewConstSummary(desc, sample.Count, sample.Sum, sample.SummaryValue, labels...)
	case metric.Histogram:
		m, err = prometheus.NewConstHistogram(desc, sample.Count, sample.Sum, sample.HistogramValue, labels...)
	default:
		m, err = prometheus.NewConstMetric(desc, getPromValueType(metricFamily.Type), sample.Value, labels...)
	}
	return m, err
}

func createMetricFamily(metrics []metric.Metric) map[string]*MetricFamily {
	mfs := make(map[string]*MetricFamily)
	for _, m := range metrics {
		tags := m.Tags()
		labels := make(map[string]string)
		for name, v := range tags {
			labels[name] = v
		}
		switch m.GetMetricType() {
		case metric.Summary:
			var name string
			var sum float64
			var count uint64
			summaryValue := make(map[float64]float64)
			for fn, fv := range m.Fields() {
				var value float64
				switch fv := fv.(type) {
				case int64:
					value = float64(fv)
				case uint64:
					value = float64(fv)
				case float64:
					value = fv
				default:
					continue
				}

				switch fn {
				case "sum":
					sum = value
				case "count":
					count = uint64(value)
				default:
					limit, err := strconv.ParseFloat(fn, 64)
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

		case metric.Histogram:
			var name string
			var sum float64
			var count uint64
			histogramValue := make(map[float64]uint64)
			for fn, fv := range m.Fields() {
				var value float64
				switch fv := fv.(type) {
				case int64:
					value = float64(fv)
				case uint64:
					value = float64(fv)
				case float64:
					value = fv
				default:
					continue
				}

				switch fn {
				case "sum":
					sum = value
				case "count":
					count = uint64(value)
				default:
					limit, err := strconv.ParseFloat(fn, 64)
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
			for fn, fv := range m.Fields() {
				var value float64
				switch fv := fv.(type) {
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
				case metric.Counter:
					if fn == "counter" {
						name = sanitize(m.GetName())
					}
				case metric.Gauge:
					if fn == "gauge" {
						name = sanitize(m.GetName())
					}
				}
				if name == "" {
					if fn == "value" {
						name = sanitize(m.GetName())
					} else {
						name = sanitize(fmt.Sprintf("%s_%s", m.GetName(), fn))
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

func addMetricFamily(mfs map[string]*MetricFamily, m metric.Metric, sample *Sample, name string) {
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

func getPromValueType(t metric.Type) prometheus.ValueType {
	switch t {
	case metric.Counter:
		return prometheus.CounterValue
	case metric.Gauge:
		return prometheus.GaugeValue
	default:
		return prometheus.UntypedValue
	}
}
