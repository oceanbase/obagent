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

package pushhttp

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/mask"
	"github.com/oceanbase/obagent/lib/slice"
	"github.com/oceanbase/obagent/lib/trace"
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/plugins/common"
	"github.com/oceanbase/obagent/stat"
)

const sampleConfig = `
protocol: prometheus
exportTimestamp: true
timestampPrecision: millisecond
batchSize: 500
taskQueueSize: 64
pushTaskCount: 8
retryTaskCount: 4
retryTimes: 0
http:
  targetAddress: http://localhost:8428
  proxyAddress: 
  apiUrl: /api/v1/import/prometheus
  httpMethod: POST
  basicAuthEnabled: true
  username: admin
  password: test
  timeout: 3s
  contentType: 'text/plain; version=0.0.4; charset=utf-8'
  headers: ["Agent-Plugin:httpOutput"]
  acceptedResponseCodes: [200, 204, 202]
  maxIdleConns: 64
  maxConnsPerHost: 64
  maxIdleConnsPerHost: 64
`

const description = `
    post metrics to any api
`

type PushHttpOutputConfig struct {
	HttpSenderConfig   HttpSenderConfig          `yaml:"http"`
	Protocol           common.Protocol           `yaml:"protocol"`
	ExportTimestamp    bool                      `yaml:"exportTimestamp"`
	TimestampPrecision common.TimestampPrecision `yaml:"timestampPrecision"`
	BatchSize          int                       `yaml:"batchSize"`
	TaskQueueSize      int                       `yaml:"taskQueueSize"`
	PushTaskCount      int                       `yaml:"pushTaskCount"`
	RetryTaskCount     int                       `yaml:"retryTaskCount"`
	RetryTimes         int                       `yaml:"retryTimes"`
	Sender
}

type MetricPushEvent struct {
	RetryTimes int
	Name       string
	Metrics    []*message.Message
	WriteTime  time.Time
}

type HttpOutput struct {
	Config                  *PushHttpOutputConfig
	prometheusCollectConfig *message.CollectorConfig
	metricEventQueue        chan *MetricPushEvent
	eventWaitGroup          sync.WaitGroup
	needStop                bool
	stopped                 chan bool
	Sender
}

func (o *HttpOutput) SampleConfig() string {
	return sampleConfig
}

func (o *HttpOutput) Description() string {
	return description
}

func (o *HttpOutput) Init(ctx context.Context, config map[string]interface{}) error {
	var pluginConfig PushHttpOutputConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "PushHttpOutput encode config")
	}
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "PushHttpOutput decode config")
	}
	o.Config = &pluginConfig

	info := fmt.Sprintf("init PushHttpOutput with config: %+v", o.Config)
	log.WithContext(ctx).Infof(mask.Mask(info))
	if o.Config.Protocol != common.Prometheus && o.Config.Protocol != common.PromeProto {
		return errors.Errorf("protocol %s is not supported.", o.Config.Protocol)
	}
	o.needStop = true
	o.stopped = make(chan bool)
	o.metricEventQueue = make(chan *MetricPushEvent, o.Config.TaskQueueSize)

	o.prometheusCollectConfig = &message.CollectorConfig{
		ExportTimestamp:    o.Config.ExportTimestamp,
		TimestampPrecision: o.Config.TimestampPrecision,
	}
	o.Config.HttpSenderConfig.retryTaskCount = o.Config.RetryTaskCount
	o.Config.HttpSenderConfig.retryTimes = o.Config.RetryTimes
	o.Sender = NewHttpSender(o.Config.HttpSenderConfig)

	for i := 0; i < o.Config.PushTaskCount; i++ {
		o.eventWaitGroup.Add(1)
		go o.doPush()
	}

	return nil
}

func (o *HttpOutput) Close() error {
	defer log.Info("http output closed")

	if !o.needStop {
		// 初始化不完整，不需要关闭资源
		return nil
	}
	o.needStop = false

	close(o.stopped)
	o.Sender.Close()
	o.eventWaitGroup.Wait()
	close(o.metricEventQueue)
	return nil
}

func (o *HttpOutput) Start(in <-chan []*message.Message) error {
	for msg := range in {
		err := o.Write(context.Background(), msg)
		if err != nil {
			log.WithError(err).Error("httpOutput write failed")
			return err
		}
	}
	return nil
}

func (o *HttpOutput) Stop() {
	defer log.Info("http output stopped")

	if !o.needStop {
		// 初始化不完整，不需要关闭资源
		return
	}
	o.needStop = false
	if o.stopped != nil {
		close(o.stopped)
	}
	if o.Sender != nil {
		o.Sender.Close()
	}
	o.eventWaitGroup.Wait()
	if o.metricEventQueue != nil {
		close(o.metricEventQueue)
	}
}

// prometheus protocol
func (o *HttpOutput) push(ctx context.Context, pushEvent *MetricPushEvent) error {
	collector := message.NewCollector(o.prometheusCollectConfig)
	collector.Fam = message.CreateMetricFamily(pushEvent.Metrics)
	registry := prometheus.NewRegistry()
	registry.Register(collector)

	mfs, err := registry.Gather()
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, o.getEncoderProto())
	for _, mf := range mfs {
		enc.Encode(mf)
	}
	return o.Sender.Send(ctx, buf.Bytes())
}

func (o *HttpOutput) getEncoderProto() expfmt.Format {
	switch o.Config.Protocol {
	case common.Prometheus:
		return expfmt.FmtText
	case common.PromeProto:
		return expfmt.FmtProtoDelim
	default:
		return expfmt.FmtText
	}
}

func (o *HttpOutput) doPush() {
	defer o.eventWaitGroup.Done()
	for {
		select {
		case _, isOpen := <-o.stopped:
			if !isOpen {
				log.Info("stop http push")
				return
			}
		case pushEvent, isOpen := <-o.metricEventQueue:
			if !isOpen {
				return
			}

			stat.HttpOutputTaskSize.With(prometheus.Labels{stat.HttpApiPath: o.Config.HttpSenderConfig.APIUrl}).Set(float64(len(o.metricEventQueue) + 1))

			ctx := trace.ContextWithRandomTraceId()
			err := o.push(ctx, pushEvent)
			if err != nil {
				log.WithContext(ctx).Warnf("push metrics %s failed, err: %s", pushEvent.Name, err)
			}
		}
	}
}

func (o *HttpOutput) Write(ctx context.Context, metrics []*message.Message) error {
	select {
	case _, isOpen := <-o.stopped:
		if !isOpen {
			log.WithContext(ctx).Info("stop http output write")
			return nil
		}
	default:
	}

	slice.SpiltBatch(len(metrics), o.Config.BatchSize, func(start, end int) {
		// Processing metric takes a long time, and closing the plug-in in the middle should respond immediately
		select {
		case <-o.stopped:
			return
		default:
		}

		stat.HttpOutputSendTaskCount.With(prometheus.Labels{stat.HttpApiPath: o.Config.HttpSenderConfig.APIUrl}).Inc()

		// copy is needed to prevent metrics from being released for a long time
		copys := make([]*message.Message, end-start)
		copy(copys, metrics[start:end])

		// If the task queue is full during the fragment process, replace the old task
		// Queue as soon as possible so that new tasks can replace the old ones
		if len(o.metricEventQueue) >= o.Config.TaskQueueSize {
			o.replaceOldTask(ctx, copys)
			return
		}

		o.metricEventQueue <- &MetricPushEvent{
			Name:      "batch",
			Metrics:   copys,
			WriteTime: time.Now(),
		}
	})

	return nil
}

func (o *HttpOutput) replaceOldTask(ctx context.Context, metrics []*message.Message) {
	log.WithContext(ctx).Warnf("push task is full, discard old task with new metrics %d", len(metrics))
	task := &MetricPushEvent{
		Name:      "batch",
		Metrics:   metrics,
		WriteTime: time.Now(),
	}

	select {
	case <-o.stopped:
		return
	case <-o.metricEventQueue:
		stat.HttpOutputTaskDiscardCount.With(prometheus.Labels{stat.HttpApiPath: o.Config.HttpSenderConfig.APIUrl}).Inc()
		o.metricEventQueue <- task
	case o.metricEventQueue <- task:
	}
}
