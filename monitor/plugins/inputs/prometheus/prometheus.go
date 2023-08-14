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

package prometheus

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/monitor/message"
)

const sampleConfig = `
addresses:['http://127.0.0.1:9090/metrics/node', 'http://127.0.0.1:9091/metrics/node']
httpTimeout: 10s
`

const description = `
collect from http server via prometheus protocol
`

var defaultTimeout = 10 * time.Second
var defaultCollectInterval = 15 * time.Second

type Config struct {
	Addresses       []string      `yaml:"addresses"`
	HttpTimeout     time.Duration `yaml:"httpTimeout"`
	CollectInterval time.Duration `yaml:"collect_interval"`
}

type Prometheus struct {
	sourceConfig map[string]interface{}

	config     *Config
	httpClient *http.Client

	ctx  context.Context
	done chan struct{}
}

func (p *Prometheus) Init(ctx context.Context, config map[string]interface{}) error {
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
	log.WithContext(ctx).Infof("prometheus input config : %v", p.config)
	p.httpClient = &http.Client{}
	if p.config.HttpTimeout == 0 {
		p.httpClient.Timeout = defaultTimeout
	} else {
		p.httpClient.Timeout = p.config.HttpTimeout
	}
	p.ctx = ctx
	p.done = make(chan struct{})
	if p.config.CollectInterval == 0 {
		p.config.CollectInterval = defaultCollectInterval
	}

	return nil
}

func (p *Prometheus) SampleConfig() string {
	return sampleConfig
}

func (p *Prometheus) Description() string {
	return description
}

func (p *Prometheus) Start(out chan<- []*message.Message) error {
	log.WithContext(p.ctx).Info("start prometheusInput")
	go p.update(p.ctx, out)
	return nil
}

func (p *Prometheus) update(ctx context.Context, out chan<- []*message.Message) {
	ticker := time.NewTicker(p.config.CollectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			msgs, err := p.CollectMsgs(ctx)
			if err != nil {
				log.WithContext(ctx).WithError(err).Warn("prometheusInput collect messages failed")
				continue
			}
			out <- msgs
		case <-p.done:
			log.Info("prometheusInput exited")
			return
		}
	}
}

func (p *Prometheus) Stop() {
	if p.done != nil {
		close(p.done)
	}
}

func (p *Prometheus) CollectMsgs(ctx context.Context) ([]*message.Message, error) {
	if p.httpClient == nil {
		return nil, errors.New("prometheus http client is nil")
	}
	var metricsTotal []*message.Message
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, address := range p.config.Addresses {
		wg.Add(1)
		go collect(ctx, p.httpClient, address, &metricsTotal, &wg, &mutex)
	}
	wg.Wait()

	return metricsTotal, nil
}

func collect(ctx context.Context, client *http.Client, url string, metricsTotal *[]*message.Message, waitGroup *sync.WaitGroup, mutex *sync.Mutex) {
	defer waitGroup.Done()

	entry := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, time.Now())).WithField("prometheus url", url)
	resp, err := client.Get(url)
	entry.Debug("get message end")
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("http client collect failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithContext(ctx).Warnf("http get resp failed status code is %d", resp.StatusCode)
		return
	}

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("read text format failed")
		return
	}

	var metrics []*message.Message
	for _, metricFamily := range metricFamilies {
		msgs := message.ParseFromMetricFamily(metricFamily)
		metrics = append(metrics, msgs...)
	}
	mutex.Lock()
	*metricsTotal = append(*metricsTotal, metrics...)
	mutex.Unlock()
}
