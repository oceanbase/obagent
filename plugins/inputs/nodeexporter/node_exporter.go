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
	"os"
	"sync"

	log2 "github.com/go-kit/kit/log/logrus"
	kitLog "github.com/go-kit/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/node_exporter/collector"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/oceanbase/obagent/metric"
)

type MetricFamilies []string

const sampleConfig = `
metricFamilies: [ node_cpu_seconds_total, node_filesystem_avail_bytes, node_filesystem_size_bytes, node_filesystem_readonly, node_disk_reads_completed_total, node_disk_read_bytes_total, node_disk_read_time_seconds_total, node_disk_writes_completed_total, node_disk_written_bytes_total, node_disk_write_time_seconds_total, node_load1, node_load15, node_load5, node_memory_Buffers_bytes, node_memory_MemFree_bytes, node_memory_Cached_bytes, node_memory_MemTotal_bytes, node_network_receive_bytes_total, node_network_transmit_bytes_total ]
`
const description = `
integrate node_exporter to collect host info
`

const (
	metricFamilies = "metricFamilies"

	nodeCpuSecondsTotal           = "node_cpu_seconds_total"
	nodeFilesystemAvailBytes      = "node_filesystem_avail_bytes"
	nodeFilesystemSizeBytes       = "node_filesystem_size_bytes"
	nodeFilesystemReadonly        = "node_filesystem_readonly"
	nodeDiskReadsCompletedTotal   = "node_disk_reads_completed_total"
	nodeDiskReadBytesTotal        = "node_disk_read_bytes_total"
	nodeDiskReadTimeSecondsTotal  = "node_disk_read_time_seconds_total"
	nodeDiskWritesCompletedTotal  = "node_disk_writes_completed_total"
	nodeDiskWrittenBytesTotal     = "node_disk_written_bytes_total"
	nodeDiskWriteTimeSecondsTotal = "node_disk_write_time_seconds_total"
	nodeLoad1                     = "node_load1"
	nodeLoad15                    = "node_load15"
	nodeLoad5                     = "node_load5"
	nodeMemoryBuffersBytes        = "node_memory_Buffers_bytes"
	nodeMemoryMemFreeBytes        = "node_memory_MemFree_bytes"
	nodeMemoryCachedBytes         = "node_memory_Cached_bytes"
	nodeMemoryMemTotalBytes       = "node_memory_MemTotal_bytes"
	nodeNetworkReceiveBytesTotal  = "node_network_receive_bytes_total"
	nodeNetworkTransmitBytesTotal = "node_network_transmit_bytes_total"
	nodeNtpOffsetSeconds          = "node_ntp_offset_seconds"
)

type NodeExporter struct {
	sourceConfig map[string]interface{}

	definedMetricFamilyNameSet map[string]int
	logger                     kitLog.Logger
	nodeCollector              *collector.NodeCollector
	registry                   *prometheus.Registry
}

func (n *NodeExporter) Init(config map[string]interface{}) error {
	n.sourceConfig = config
	n.definedMetricFamilyNameSet = make(map[string]int, 32)
	if _, ok := n.sourceConfig[metricFamilies]; !ok {
		return errors.New("node exporter metric families is not exist")
	}
	families, ok := n.sourceConfig[metricFamilies].([]interface{})
	if !ok {
		return errors.New("node exporter metric families sourceConfig error")
	}

	for _, familyName := range families {
		switch familyName {
		case nodeCpuSecondsTotal:
			n.definedMetricFamilyNameSet[nodeCpuSecondsTotal] = 1

		case nodeFilesystemAvailBytes:
			n.definedMetricFamilyNameSet[nodeFilesystemAvailBytes] = 1

		case nodeFilesystemSizeBytes:
			n.definedMetricFamilyNameSet[nodeFilesystemSizeBytes] = 1

		case nodeFilesystemReadonly:
			n.definedMetricFamilyNameSet[nodeFilesystemReadonly] = 1

		case nodeDiskReadsCompletedTotal:
			n.definedMetricFamilyNameSet[nodeDiskReadsCompletedTotal] = 1

		case nodeDiskReadBytesTotal:
			n.definedMetricFamilyNameSet[nodeDiskReadBytesTotal] = 1

		case nodeDiskReadTimeSecondsTotal:
			n.definedMetricFamilyNameSet[nodeDiskReadTimeSecondsTotal] = 1

		case nodeDiskWritesCompletedTotal:
			n.definedMetricFamilyNameSet[nodeDiskWritesCompletedTotal] = 1

		case nodeDiskWrittenBytesTotal:
			n.definedMetricFamilyNameSet[nodeDiskWrittenBytesTotal] = 1

		case nodeDiskWriteTimeSecondsTotal:
			n.definedMetricFamilyNameSet[nodeDiskWriteTimeSecondsTotal] = 1

		case nodeLoad1:
			n.definedMetricFamilyNameSet[nodeLoad1] = 1

		case nodeLoad15:
			n.definedMetricFamilyNameSet[nodeLoad15] = 1

		case nodeLoad5:
			n.definedMetricFamilyNameSet[nodeLoad5] = 1

		case nodeMemoryBuffersBytes:
			n.definedMetricFamilyNameSet[nodeMemoryBuffersBytes] = 1

		case nodeMemoryMemFreeBytes:
			n.definedMetricFamilyNameSet[nodeMemoryMemFreeBytes] = 1

		case nodeMemoryCachedBytes:
			n.definedMetricFamilyNameSet[nodeMemoryCachedBytes] = 1

		case nodeMemoryMemTotalBytes:
			n.definedMetricFamilyNameSet[nodeMemoryMemTotalBytes] = 1

		case nodeNetworkReceiveBytesTotal:
			n.definedMetricFamilyNameSet[nodeNetworkReceiveBytesTotal] = 1

		case nodeNetworkTransmitBytesTotal:
			n.definedMetricFamilyNameSet[nodeNetworkTransmitBytesTotal] = 1

		case nodeNtpOffsetSeconds:
			n.definedMetricFamilyNameSet[nodeNtpOffsetSeconds] = 1

		default:
			return errors.Errorf("node exporter metric families %s is undefined", familyName)
		}
	}

	commandLineParse()

	n.logger = log2.NewLogrusLogger(log.StandardLogger())

	var err error
	n.nodeCollector, err = collector.NewNodeCollector(n.logger)
	if err != nil {
		return errors.Wrap(err, "node exporter create collector")
	}

	n.registry = prometheus.NewRegistry()
	err = n.registry.Register(n.nodeCollector)
	if err != nil {
		return errors.Wrap(err, "node exporter register collector")
	}

	return nil
}

var once sync.Once

func commandLineParse() {
	once.Do(func() {

		lastIndex := len(os.Args) - 1
		copy(os.Args[lastIndex:], os.Args)
		os.Args = os.Args[lastIndex:]

		for _, item := range collectItems {
			kingpin.CommandLine.GetFlag(item).Default("true")
		}

		for _, item := range noCollectItems {
			kingpin.CommandLine.GetFlag(item).Default("false")
		}

		kingpin.Parse()
	})
}

func (n *NodeExporter) Close() error {
	n.registry.Unregister(n.nodeCollector)
	return nil
}

func (n *NodeExporter) SampleConfig() string {
	return sampleConfig
}

func (n *NodeExporter) Description() string {
	return description
}

func (n *NodeExporter) Collect() ([]metric.Metric, error) {
	var metrics []metric.Metric

	metricFamilies, err := n.registry.Gather()
	if err != nil {
		return nil, errors.Wrap(err, "node exporter registry gather")
	}
	for _, metricFamily := range metricFamilies {

		_, exist := n.definedMetricFamilyNameSet[metricFamily.GetName()]
		if exist {

			metricsFromMetricFamily := metric.ParseFromMetricFamily(metricFamily)
			metrics = append(metrics, metricsFromMetricFamily...)
		}
	}

	return metrics, nil
}
