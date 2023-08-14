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

package label

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/monitor/message"
)

func newTestMetric() *message.Message {
	name := "test"

	metricType := message.Gauge

	currentTime := time.Now()

	return message.NewMessage(name, metricType, currentTime).
		AddTag("mountpoint", "/home").
		AddField("f1", 1.0).AddField("f2", 2.0)

}

func newTestMetrics() []*message.Message {
	metricEntry := newTestMetric()
	var metrics []*message.Message
	metrics = append(metrics, metricEntry)
	return metrics
}

func TestGetMountPath(t *testing.T) {
	metrics := newTestMetrics()
	configStr := `
      labelTags:
        installPath: /home/admin/oceanbase
        dataDiskPath: /data/1
        logDiskPath: /data/log1
    `
	var mountLabelConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &mountLabelConfigMap)

	mountLabelProcessor := &MountLabelProcessor{}
	mountLabelProcessor.Init(context.Background(), mountLabelConfigMap)
	metricsProcessed, _ := mountLabelProcessor.Process(context.Background(), metrics...)
	require.Equal(t, 1, len(metricsProcessed))
}
