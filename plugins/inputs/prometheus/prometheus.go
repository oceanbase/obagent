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
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
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
	for _, metricFamily := range metricFamilies {
		metricsFromMetricFamily := metric.ParseFromMetricFamily(metricFamily)
		metrics = append(metrics, metricsFromMetricFamily...)
	}
	mutex.Lock()
	*metricsTotal = append(*metricsTotal, metrics...)
	mutex.Unlock()
}
