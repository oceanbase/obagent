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
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/utils"
)

// implete interface Sender
type printer struct {
	receivedCount int32
	fail          bool
}

func (p *printer) Close() { p.receivedCount = -1 }

func (p *printer) setFail(fail bool) {
	p.fail = false
}

func (p *printer) Send(ctx context.Context, data []byte) error {
	atomic.AddInt32(&p.receivedCount, 1)
	log.Infof("receive metric data size:%d", len(data))
	if p.fail {
		return errors.New("fail")
	}
	return nil
}

func newTestMetrics() []*message.Message {
	var metrics = make([]*message.Message, 0)
	var numStrs = []string{"a", "b", "c"}
	for _, numStr := range numStrs {
		metricName := "testMetric" + "_" + numStr
		metricType := message.Gauge
		for i := 0; i < 3; i++ {
			si := strconv.Itoa(i)
			t := time.Now()
			metricEntry := message.NewMessage(metricName, metricType, t).
				AddTag("host", "app").
				AddTag("tag1", si).
				AddField("value", 10.0)
			metrics = append(metrics, metricEntry)
		}
	}
	return metrics
}

func TestInit(t *testing.T) {
	pushhttpOutput := HttpOutput{}
	defer pushhttpOutput.Close()
	config := `
      protocol: prometheus
      exportTimestamp: true
      timestampPrecision: millisecond
      batchSize: 500
      taskQueueSize: 64
      pushTaskCount: 8
      retryTaskCount: 4
      retryTimes: 1
      http:
        targetAddress: http://localhost:8088
        proxyAddress: 
        apiUrl: /api/v1/metrics/import
        httpMethod: POST
        basicAuthEnabled: false
        username: 
        password: 
        timeout: 1s
        contentType: 'text/plain; version=0.0.4; charset=utf-8'
        headers: ["key1:value1", "key2:value2"]
        acceptedResponseCodes: [200, 202]
    `
	configMap, _ := utils.DecodeYaml(config)
	err := pushhttpOutput.Init(context.Background(), configMap)
	assert.Nil(t, err)
}

func TestWriteWithHttpSenderFail(t *testing.T) {
	pushhttpOutput := HttpOutput{}
	config := `
      protocol: prometheus
      exportTimestamp: true
      timestampPrecision: millisecond
      batchSize: 1
      taskQueueSize: 64
      pushTaskCount: 8
      retryTaskCount: 4
      retryTimes: 10
      http:
        targetAddress: http://localhost:65535
        proxyAddress: 
        apiUrl: /api/v1/metrics/import
        httpMethod: POST
        basicAuthEnabled: false
        username: 
        password: 
        timeout: 1s
        contentType: 'text/plain; version=0.0.4; charset=utf-8'
        headers: ["key1:value1", "key2:value2"]
        acceptedResponseCodes: [200, 202]
    `
	configMap, _ := utils.DecodeYaml(config)
	err := pushhttpOutput.Init(context.Background(), configMap)
	assert.Nil(t, err)

	metrics := newTestMetrics()
	pushhttpOutput.Write(context.Background(), metrics)

	time.Sleep(time.Second * 2)

	pushhttpOutput.Close()
}

func TestWriteOnce(t *testing.T) {
	pushhttpOutput := HttpOutput{}
	config := `
    protocol: prometheus
    exportTimestamp: true
    timestampPrecision: millisecond
    batchSize: 500
    taskQueueSize: 64
    pushTaskCount: 8
    retryTaskCount: 4
    retryTimes: 1`
	configMap, _ := utils.DecodeYaml(config)
	err := pushhttpOutput.Init(context.Background(), configMap)
	assert.Nil(t, err)

	printer := new(printer)
	pushhttpOutput.Sender = printer

	metrics := newTestMetrics()
	pushhttpOutput.Write(context.Background(), metrics)
	time.Sleep(time.Second)
	assert.Equal(t, int32(1), printer.receivedCount)

	pushhttpOutput.Close()
	assert.Equal(t, int32(-1), printer.receivedCount)
}

func TestWriteOnce_Batch_1(t *testing.T) {
	pushhttpOutput := HttpOutput{}
	config := `
    protocol: prometheus
    exportTimestamp: true
    timestampPrecision: millisecond
    batchSize: 1
    taskQueueSize: 64
    pushTaskCount: 8
    retryTaskCount: 4
    retryTimes: 1`
	configMap, _ := utils.DecodeYaml(config)
	err := pushhttpOutput.Init(context.Background(), configMap)
	assert.Nil(t, err)

	printer := new(printer)
	pushhttpOutput.Sender = printer

	metrics := newTestMetrics()
	pushhttpOutput.Write(context.Background(), metrics)
	time.Sleep(time.Second * 1)
	assert.Equal(t, int32(len(metrics)), printer.receivedCount)

	pushhttpOutput.Close()
	assert.Equal(t, int32(-1), printer.receivedCount)
}
