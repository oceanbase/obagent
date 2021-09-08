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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/oceanbase/obagent/metric"
)

func TestPrometheus_Export(t *testing.T) {
	config := `{"formatType":"fmtText"}`
	sourceConfig := make(map[string]interface{})
	err := json.Unmarshal([]byte(config), &sourceConfig)
	if err != nil {
		t.Errorf("json Unmarshal err %s", err.Error())
	}
	p := &Prometheus{}
	err = p.Init(sourceConfig)
	if err != nil {
		t.Errorf("exporter prometheus init err %s", err.Error())
	}
	var metricTypes = []metric.Type{metric.Counter, metric.Gauge, metric.Summary, metric.Histogram, metric.Untyped}
	metrics := make([]metric.Metric, 0, 100)
	for i := 0; i < 100; i++ {
		metrics = append(metrics,
			metric.NewMetric(
				fmt.Sprintf("node_disk_written_sectors_total%d", i),
				map[string]interface{}{"counter": 651.091796875},
				map[string]string{"device": "disk0"},
				time.Now(),
				metricTypes[i%len(metricTypes)]),
		)
	}

	type args struct {
		metrics []metric.Metric
	}
	tests := []struct {
		name    string
		fields  Prometheus
		args    args
		wantErr bool
	}{
		{name: "test1", fields: *p, args: args{metrics: metrics}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Prometheus{
				sourceConfig: tt.fields.sourceConfig,
				config:       tt.fields.config,
			}
			buffer, err := p.Export(tt.args.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("Export() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("exporter prometheus test export result: %s", buffer.String())
		})
	}
}
