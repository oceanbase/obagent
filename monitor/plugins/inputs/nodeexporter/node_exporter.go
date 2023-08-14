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

package nodeexporter

import (
	"context"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	log2 "github.com/go-kit/kit/log/logrus"
	kitLog "github.com/go-kit/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/node_exporter/collector"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/monitor/message"
)

type MetricFamilies []string

const sampleConfig = `
collectors: [ cpu, diskstats, loadavg, meminfo, filesystem, netdev ]
metricFamilies: [ node_cpu_seconds_total, node_filesystem_avail_bytes, node_filesystem_size_bytes, node_filesystem_readonly, node_disk_reads_completed_total, node_disk_read_bytes_total, node_disk_read_time_seconds_total, node_disk_writes_completed_total, node_disk_written_bytes_total, node_disk_write_time_seconds_total, node_load1, node_load15, node_load5, node_memory_Buffers_bytes, node_memory_MemFree_bytes, node_memory_Cached_bytes, node_memory_MemTotal_bytes, node_network_receive_bytes_total, node_network_transmit_bytes_total ]
`
const description = `
integrate node_exporter to collect host info
`

const (
	metricFamilies = "metricFamilies"
	collectors     = "collectors"

	nodeCpuSecondsTotal                = "node_cpu_seconds_total"
	nodeFilesystemAvailBytes           = "node_filesystem_avail_bytes"
	nodeFilesystemSizeBytes            = "node_filesystem_size_bytes"
	nodeFilesystemReadonly             = "node_filesystem_readonly"
	nodeDiskReadsCompletedTotal        = "node_disk_reads_completed_total"
	nodeDiskReadBytesTotal             = "node_disk_read_bytes_total"
	nodeDiskReadTimeSecondsTotal       = "node_disk_read_time_seconds_total"
	nodeDiskWritesCompletedTotal       = "node_disk_writes_completed_total"
	nodeDiskWrittenBytesTotal          = "node_disk_written_bytes_total"
	nodeDiskWriteTimeSecondsTotal      = "node_disk_write_time_seconds_total"
	nodeLoad1                          = "node_load1"
	nodeLoad15                         = "node_load15"
	nodeLoad5                          = "node_load5"
	nodeMemoryBuffersBytes             = "node_memory_Buffers_bytes"
	nodeMemoryMemFreeBytes             = "node_memory_MemFree_bytes"
	nodeMemoryCachedBytes              = "node_memory_Cached_bytes"
	nodeMemoryMemTotalBytes            = "node_memory_MemTotal_bytes"
	nodeNetworkReceiveBytesTotal       = "node_network_receive_bytes_total"
	nodeNetworkTransmitBytesTotal      = "node_network_transmit_bytes_total"
	nodeNtpOffsetSeconds               = "node_ntp_offset_seconds"
	nodeFilesystemFilesFree            = "node_filesystem_files_free"
	nodeFilesystemFiles                = "node_filesystem_files"
	nodeNetworkReceiveErrsTotal        = "node_network_receive_errs_total"
	nodeNetworkTransmitErrsTotal       = "node_network_transmit_errs_total"
	nodeNetworkReceiveDropTotal        = "node_network_receive_drop_total"
	nodeNetworkTransmitDropTotal       = "node_network_transmit_drop_total"
	nodeDiskIoTimeWeightedSecondsTotal = "node_disk_io_time_weighted_seconds_total"

	cpu        = "cpu"
	diskstats  = "diskstats"
	netdev     = "netdev"
	filesystem = "filesystem"
	loadavg    = "loadavg"
	meminfo    = "meminfo"
)

var defaultCollectors = []string{cpu, diskstats, netdev, filesystem, loadavg, meminfo}

type NodeExporter struct {
	sourceConfig map[string]interface{}

	definedMetricFamilyNameSet map[string]int
	definedCollectorSet        map[string]int
	logger                     kitLog.Logger
	nodeCollector              *collector.NodeCollector
	registryMap                map[string]*prometheus.Registry
	CollectInterval            time.Duration

	ctx  context.Context
	done chan struct{}
}

func (n *NodeExporter) Init(ctx context.Context, config map[string]interface{}) error {
	n.sourceConfig = config
	n.ctx = context.Background()
	n.done = make(chan struct{})
	n.definedMetricFamilyNameSet = make(map[string]int, 32)
	if _, ok := n.sourceConfig[metricFamilies]; !ok {
		return errors.New("node exporter message families is not exist")
	}
	families, ok := n.sourceConfig[metricFamilies].([]interface{})
	if !ok {
		return errors.New("node exporter message families sourceConfig error")
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

		case nodeFilesystemFilesFree:
			n.definedMetricFamilyNameSet[nodeFilesystemFilesFree] = 1

		case nodeFilesystemFiles:
			n.definedMetricFamilyNameSet[nodeFilesystemFiles] = 1

		case nodeNetworkReceiveErrsTotal:
			n.definedMetricFamilyNameSet[nodeNetworkReceiveErrsTotal] = 1

		case nodeNetworkTransmitErrsTotal:
			n.definedMetricFamilyNameSet[nodeNetworkTransmitErrsTotal] = 1

		case nodeNetworkReceiveDropTotal:
			n.definedMetricFamilyNameSet[nodeNetworkReceiveDropTotal] = 1

		case nodeNetworkTransmitDropTotal:
			n.definedMetricFamilyNameSet[nodeNetworkTransmitDropTotal] = 1

		case nodeDiskIoTimeWeightedSecondsTotal:
			n.definedMetricFamilyNameSet[nodeDiskIoTimeWeightedSecondsTotal] = 1

		default:
			return errors.Errorf("node exporter message families %s is undefined", familyName)
		}
	}

	n.definedCollectorSet = make(map[string]int, 16)
	collectors, ok := n.sourceConfig[collectors].([]interface{})
	if !ok {
		log.WithContext(ctx).Infof("node_exporter collectors config is empty")
	}
	if len(collectors) > 0 {
		for _, c := range collectors {
			switch c {
			case cpu:
				n.definedCollectorSet[cpu] = 1
			case diskstats:
				n.definedCollectorSet[diskstats] = 1
			case netdev:
				n.definedCollectorSet[netdev] = 1
			case filesystem:
				n.definedCollectorSet[filesystem] = 1
			case loadavg:
				n.definedCollectorSet[loadavg] = 1
			case meminfo:
				n.definedCollectorSet[meminfo] = 1
			default:
				return errors.Errorf("node exporter collector %s is undefined", c)
			}
		}
	} else {
		for _, c := range defaultCollectors {
			n.definedCollectorSet[c] = 1
		}
	}
	interval, ok := n.sourceConfig["collect_interval"].(time.Duration)
	if !ok {
		log.WithContext(ctx).Warnf("node_exporter collect_interval config is not correct, using default config")
		interval = 15 * time.Second
	}
	n.CollectInterval = interval

	commandLineParse()

	n.logger = log2.NewLogrusLogger(log.StandardLogger())

	var registryMap = make(map[string]*prometheus.Registry)
	for c := range n.definedCollectorSet {
		nodeCollector, err := collector.NewNodeCollector(n.logger, c)
		if err != nil {
			return errors.Wrap(err, "node exporter create collector")
		}
		registry := prometheus.NewRegistry()
		err = registry.Register(nodeCollector)
		if err != nil {
			return errors.Wrap(err, "node exporter register collector")
		}
		registryMap[c] = registry
	}
	n.registryMap = registryMap

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

func (n *NodeExporter) SampleConfig() string {
	return sampleConfig
}

func (n *NodeExporter) Description() string {
	return description
}

func (n *NodeExporter) Start(out chan<- []*message.Message) error {
	log.Info("start nodeExporter")
	go n.update(n.ctx, out)
	return nil
}

func (n *NodeExporter) update(ctx context.Context, out chan<- []*message.Message) {
	ticker := time.NewTicker(n.CollectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			msgs, err := n.CollectMsgs(ctx)
			if err != nil {
				log.WithContext(ctx).Warnf("node_exporter collect failed, reason: %s", err)
			}
			out <- msgs
		case <-n.done:
			log.WithContext(ctx).Infof("node_exporter exited")
			return
		}
	}
}

func (n *NodeExporter) Stop() {
	if n.done != nil {
		close(n.done)
	}
}

func (n *NodeExporter) doCollect(ctx context.Context, name string, registry *prometheus.Registry) ([]*message.Message, error) {
	var metrics []*message.Message
	entry := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, time.Now())).WithField("input", "nodeexporter")
	metricFamilies, err := registry.Gather()
	entry.Debug("gather end")
	if err != nil {
		return nil, errors.Wrapf(err, "node exporter %s registry gather", name)
	}
	now := time.Now()
	for _, metricFamily := range metricFamilies {

		_, exist := n.definedMetricFamilyNameSet[metricFamily.GetName()]
		if exist {

			for _, m := range metricFamily.Metric {
				tags := makeLabels(m)
				var fields []message.FieldEntry

				switch metricFamily.GetType() {

				case dto.MetricType_SUMMARY:
					fields = makeQuantiles(m)
					fields = append(fields,
						message.FieldEntry{Name: "count", Value: float64(m.GetSummary().GetSampleCount())},
						message.FieldEntry{Name: "sum", Value: m.GetSummary().GetSampleSum()})
				case dto.MetricType_HISTOGRAM:
					fields = makeBuckets(m)
					fields = append(fields,
						message.FieldEntry{Name: "count", Value: float64(m.GetHistogram().GetSampleCount())},
						message.FieldEntry{Name: "sum", Value: m.GetHistogram().GetSampleSum()})
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
					newMetric := message.NewMessageWithTagsFields(metricFamily.GetName(), ValueType(metricFamily.GetType()), t, tags, fields)
					metrics = append(metrics, newMetric)
				}
			}
		}
	}

	return metrics, nil
}

func (n *NodeExporter) CollectMsgs(ctx context.Context) ([]*message.Message, error) {
	metrics := make([]*message.Message, 0)

	for c, registry := range n.registryMap {
		msgs, err := n.doCollect(ctx, c, registry)
		if err != nil {
			log.WithContext(ctx).Warnf("collect node exporter %s failed, reason: %s", c, err)
			continue
		}
		metrics = append(metrics, msgs...)
	}
	return metrics, nil
}

func makeLabels(m *dto.Metric) []message.TagEntry {
	result := make([]message.TagEntry, 0, len(m.Label))
	for _, lp := range m.Label {
		result = append(result, message.TagEntry{Name: lp.GetName(), Value: lp.GetValue()})
	}
	return result
}

func makeQuantiles(m *dto.Metric) []message.FieldEntry {
	fields := make([]message.FieldEntry, 0, len(m.GetSummary().Quantile)+2)
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			fields = append(fields, message.FieldEntry{Name: fmt.Sprint(q.GetQuantile()), Value: q.GetValue()})
		}
	}
	return fields
}

func makeBuckets(m *dto.Metric) []message.FieldEntry {
	fields := make([]message.FieldEntry, 0, len(m.GetHistogram().Bucket)+2)
	for _, b := range m.GetHistogram().Bucket {
		fields = append(fields, message.FieldEntry{Name: fmt.Sprint(b.GetUpperBound()), Value: float64(b.GetCumulativeCount())})
	}
	return fields
}

func getNameAndValue(m *dto.Metric) []message.FieldEntry {
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			return []message.FieldEntry{{Name: "gauge", Value: m.GetGauge().GetValue()}}
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			return []message.FieldEntry{{Name: "counter", Value: m.GetCounter().GetValue()}}
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			return []message.FieldEntry{{Name: "value", Value: m.GetUntyped().GetValue()}}
		}
	}
	return []message.FieldEntry{}
}

func ValueType(metricType dto.MetricType) message.Type {
	switch metricType {
	case dto.MetricType_COUNTER:
		return message.Counter
	case dto.MetricType_GAUGE:
		return message.Gauge
	case dto.MetricType_SUMMARY:
		return message.Summary
	case dto.MetricType_HISTOGRAM:
		return message.Histogram
	default:
		return message.Untyped
	}
}
