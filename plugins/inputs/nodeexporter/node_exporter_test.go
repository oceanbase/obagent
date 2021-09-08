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

package nodeexporter

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/oceanbase/obagent/metric"
)

func TestNodeExporter_Collect(t *testing.T) {
	lastIndex := len(os.Args) - 1
	copy(os.Args[lastIndex:], os.Args)
	os.Args = os.Args[lastIndex:]

	config := `
{
	"metricFamilies": ["node_cpu_seconds_total", "node_filesystem_avail_bytes",
		"node_filesystem_size_bytes",
		"node_filesystem_readonly",
		"node_disk_reads_completed_total",
		"node_disk_read_bytes_total",
		"node_disk_read_time_seconds_total",
		"node_disk_writes_completed_total",
		"node_disk_written_bytes_total",
		"node_disk_write_time_seconds_total",
		"node_load1",
		"node_load15",
		"node_load5",
		"node_memory_Buffers_bytes",
		"node_memory_MemFree_bytes",
		"node_memory_Cached_bytes",
		"node_memory_MemTotal_bytes",
		"node_network_receive_bytes_total",
		"node_network_transmit_bytes_total",
		"node_ntp_offset_seconds"
	]
}
`
	n := &NodeExporter{}
	sourceConfig := make(map[string]interface{})
	err := json.Unmarshal([]byte(config), &sourceConfig)
	if err != nil {
		t.Errorf("json Unmarshal err %s", err.Error())
	}
	err = n.Init(sourceConfig)
	if err != nil {
		t.Errorf("node exporter init err %s", err.Error())
	}

	tests := []struct {
		name    string
		fields  NodeExporter
		want    []metric.Metric
		wantErr bool
	}{
		{name: "test1", fields: *n, want: nil, wantErr: false},
		{name: "test2", fields: *n, want: nil, wantErr: false},
		{name: "test3", fields: *n, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeExporter{
				sourceConfig:               tt.fields.sourceConfig,
				definedMetricFamilyNameSet: tt.fields.definedMetricFamilyNameSet,
				logger:                     tt.fields.logger,
				nodeCollector:              tt.fields.nodeCollector,
				registry:                   tt.fields.registry,
			}
			_, err := n.Collect()
			if (err != nil) != tt.wantErr {
				t.Errorf("Collect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
