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
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/metric"
)

const sampleConfig = `
addresses:['http://127.0.0.1:9090/metrics/node', 'http://127.0.0.1:9091/metrics/node']
httpTimeout: 10s
`

const description = `
collect from http server via prometheus protocol
`

var defaultTimeout = 10 * time.Second

type Config struct {
	Addresses   []string      `yaml:"addresses"`
	HttpTimeout time.Duration `yaml:"httpTimeout"`
}

type Prometheus struct {
	sourceConfig map[string]interface{}

	config     *Config
	httpClient *http.Client
}

func (p *Prometheus) Init(config map[string]interface{}) error {
	p.sourceConfig = config
	configData, err := yaml.Marshal(p.sourceConfig)
	if err != nil {
		return errors.Wrap(err, "prometheus input encode config")
	}
	p.config = &Config{}
	err = yaml.Unmarshal(configData, p.config)
	if err != nil {
		return errors.Wrap(err, "prometheus input decode config")
	}
	log.Infof("prometheus input config : %v", p.config)
	p.httpClient = &http.Client{}
	if p.config.HttpTimeout == 0 {
		p.httpClient.Timeout = defaultTimeout
	} else {
		p.httpClient.Timeout = p.config.HttpTimeout
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

func (p *Prometheus) Collect() ([]metric.Metric, error) {
	if p.httpClient == nil {
		return nil, errors.New("prometheus http client is nil")
	}
	var metricsTotal []metric.Metric
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, address := range p.config.Addresses {
		wg.Add(1)
		go collect(p.httpClient, address, &metricsTotal, &wg, &mutex)
	}
	wg.Wait()

	return metricsTotal, nil
}

func collect(client *http.Client, url string, metricsTotal *[]metric.Metric, waitGroup *sync.WaitGroup, mutex *sync.Mutex) {
	defer waitGroup.Done()

	resp, err := client.Get(url)
	if err != nil {
		log.WithError(err).Error("http client collect failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("http get resp failed status code is %d", resp.StatusCode)
		return
	}

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		log.WithError(err).Error("read text format failed")
		return
	}

	var metrics []metric.Metric
	now := time.Now()
	for _, metricFamily := range metricFamilies {

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
				newMetric := metric.NewMetric(metricFamily.GetName(), fields, tags, t, ValueType(metricFamily.GetType()))

				metrics = append(metrics, newMetric)
			}

		}

	}
	mutex.Lock()
	*metricsTotal = append(*metricsTotal, metrics...)
	mutex.Unlock()
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

func ValueType(metricType dto.MetricType) metric.Type {
	switch metricType {
	case dto.MetricType_COUNTER:
		return metric.Counter
	case dto.MetricType_GAUGE:
		return metric.Gauge
	case dto.MetricType_SUMMARY:
		return metric.Summary
	case dto.MetricType_HISTOGRAM:
		return metric.Histogram
	default:
		return metric.Untyped
	}
}
