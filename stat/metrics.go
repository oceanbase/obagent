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

package stat

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
)

var (
	defaultGatherer prometheus.Gatherer = prometheus.DefaultGatherer
)

func RegisterStat(ctx context.Context) {
	registry := prometheus.NewRegistry()
	labels := prometheus.Labels{
		App:     Host,
		Process: config.CurProcess,
		SvrIP:   HostIP,
	}
	log.WithContext(ctx).Infof("registry with labels: %+v", labels)
	registerer := prometheus.WrapRegistererWith(labels, registry)

	registerer.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
		HttpRequestMillisecondsSummary,
		MonAgentPipelineReportMetricsTotal,
		MonAgentPipelineExecuteTotal,
		MonAgentPipelineExecuteSecondsTotal,
		MonAgentPluginExecuteTotal,
		MonAgentPluginExecuteSecondsTotal,
		MonitorAgentTableInputHistogram,
		InputCollectMetricsTotal,
		ProcessorProcessMetricsTotal,
		OutputWriteMetricsTotal,
		ExporterExportMetricsTotal,
		ExporterExportMetricsBytesTotal,
		HttpOutputPushMetricsBytesTotal,
		HttpOutputPushTotal,
		HttpOutputTaskSize,
		HttpOutputRetryTaskSize,
		HttpOutputTaskDiscardCount,
		HttpOutputSendTaskCount,
		HttpOutputSendFailedCount,
		HttpOutputSendRetryCount,
		HttpOutputSendMillisecondsSummary,
		SqlAuditInputCollectDataJumpedCount,
		MysqlOutputWriteTaskSize,
		MysqlOutputWriteSqlCount,
		MysqlOutputWriteSqlFailedCount,
		MysqlOutputContextTimeoutDiscardMetricsCount,
		MysqlOutputWriteSqlRetryCount,
		MysqlOutputWriteSqlMillisecondsSummary,
		LogTailerCount,
		LogTailerTailLineCount,
		LogTailerReadingFileOffset,
		LogTailerReadingFileId,
		LogTailerProcessQueueSize,
	)

	gatherPtr, _ := defaultGatherer.(*prometheus.Registry)
	*gatherPtr = *registry
}

var (

	// HttpRequestMillisecondsSummary gtto request seconds total
	HttpRequestMillisecondsSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_milliseconds_summary",
			Help: "The total elapsed time in millisecond of http request",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		}, []string{HttpMethod, HttpStatus, HttpApiPath},
	)

	//MonAgentPipelineBufferMetrics monitor pipeline buffer metrics
	_ = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monagent_pipeline_buffer_metrics",
		Help: "The number of metrics currently in pipeline output buffer",
	}, []string{PluginNameKey})

	//MonAgentPipelineReportMetricsTotal monitor pipeline report message total
	MonAgentPipelineReportMetricsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_pipeline_report_metrics_total",
		Help: "The total number of metrics reported from the pipeline",
	}, []string{PluginNameKey})

	//MonAgentPipelineExecuteTotal monitor pipeline execute total
	MonAgentPipelineExecuteTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_pipeline_execute_total",
		Help: "The total number of a pipeline executed",
	}, []string{PluginNameKey, PluginExecuteStatusKey})

	//MonAgentPipelineExecuteSecondsTotal monitor pipeline execute seconds total
	MonAgentPipelineExecuteSecondsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_pipeline_execute_seconds_total",
		Help: "The total time in second of a pipeline executed",
	}, []string{PluginNameKey, PluginExecuteStatusKey})

	//MonAgentPluginExecuteTotal monitor plugin execute total
	MonAgentPluginExecuteTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_plugin_execute_total",
		Help: "The total number of a plugin executed",
	}, []string{PluginNameKey, PluginExecuteStatusKey, PluginTypeKey})

	//MonAgentPluginExecuteSecondsTotal monitor plugin execute seconds total
	MonAgentPluginExecuteSecondsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_plugin_execute_seconds_total",
		Help: "The total time in second of a plugin executed",
	}, []string{PluginNameKey, PluginExecuteStatusKey, PluginTypeKey})

	MonitorAgentTableInputHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "monagent",
			Subsystem: "input",
			Name:      "sql_duration_seconds",
			Help:      "Bucketed histogram of execute time (s) of collecting sql",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1},
		}, []string{PluginNameKey})

	InputCollectMetricsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "input_collect_metrics_total",
		Help: "The total metrics count that input collected",
	}, []string{PluginNameKey, PluginTypeKey})

	ProcessorProcessMetricsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "processor_process_metrics_total",
		Help: "The total metrics count that processor processed",
	}, []string{PluginNameKey, PluginTypeKey})

	OutputWriteMetricsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "output_write_metrics_total",
		Help: "The total metrics count that output written",
	}, []string{PluginNameKey, PluginTypeKey})

	ExporterExportMetricsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "exporter_export_metrics_total",
		Help: "The total metrics that exporter export",
	}, []string{PluginNameKey, PluginTypeKey, HttpApiPath})

	ExporterExportMetricsBytesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "exporter_export_metrics_bytes_total",
		Help: "The total metrics bytes that exporter export",
	}, []string{PluginNameKey, PluginTypeKey, HttpApiPath})

	HttpOutputPushMetricsBytesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_output_push_metrics_bytes_total",
		Help: "The total bytes of metrics that http output pushed",
	}, []string{HttpApiPath})

	HttpOutputPushTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_output_push_total",
		Help: "http output push count",
	}, []string{HttpApiPath})

	HttpOutputTaskSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_output_task_size",
		Help: "http output task size",
	}, []string{HttpApiPath})

	HttpOutputRetryTaskSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_output_retry_task_size",
		Help: "http output retry task size",
	}, []string{HttpApiPath})

	HttpOutputTaskDiscardCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_output_task_discard_total",
		Help: "http output discard task count",
	}, []string{HttpApiPath})

	HttpOutputSendTaskCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_output_send_task_total",
		Help: "http output send task count",
	}, []string{HttpApiPath})

	HttpOutputSendFailedCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_output_send_failed_total",
		Help: "http output send failed count",
	}, []string{HttpApiPath})

	HttpOutputSendRetryCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_output_send_retry_total",
		Help: "http output send retry count",
	}, []string{HttpApiPath})

	// contains retry data
	HttpOutputSendMillisecondsSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_output_send_milliseconds_summary",
			Help: "The total elapsed time of http output",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		}, []string{HttpMethod, HttpStatus, HttpApiPath},
	)

	SqlAuditInputCollectDataJumpedCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sql_audit_input_collect_data_jumped_total",
		Help: "sql audit input plugin collect data jumped total",
	}, []string{PluginNameKey})

	MysqlOutputWriteTaskSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mysql_output_write_task_size",
		Help: "mysql output write task chan size",
	}, []string{MysqlOutputTaskNameKey})

	MysqlOutputWriteSqlCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mysql_output_write_sql_count",
		Help: "mysql output write sql total",
	}, []string{MysqlOutputTableNameKey})

	MysqlOutputWriteSqlFailedCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mysql_output_write_sql_failed_count",
		Help: "mysql output write sql failed total",
	}, []string{MysqlOutputTableNameKey})

	MysqlOutputContextTimeoutDiscardMetricsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mysql_output_context_timeout_discard_metrics_count",
		Help: "mysql output context timeout discard metrics total",
	}, []string{MysqlOutputMetricName})

	MysqlOutputWriteSqlRetryCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mysql_output_write_sql_retry_count",
		Help: "mysql output write sql retry total",
	}, []string{MysqlOutputTableNameKey})

	// MysqlOutputWriteSqlMillisecondsSummary write sql seconds total
	MysqlOutputWriteSqlMillisecondsSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "mysql_output_write_sql_milliseconds_summary",
			Help: "The total elapsed time in millisecond of mysql output write sql",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		}, []string{MysqlOutputTableNameKey},
	)

	LogTailerCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "log_tailer_count",
		Help: "log tailer count",
	}, []string{LogFileName})
	LogTailerTailLineCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "log_tailer_tail_line_count",
		Help: "log tailer tail line count",
	}, []string{LogFileName})
	LogTailerReadingFileOffset = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "log_tailer_reading_file_offset",
		Help: "log tailer reading file offset",
	}, []string{LogFileName})
	LogTailerReadingFileId = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "log_tailer_reading_file_id",
		Help: "log tailer reading file id",
	}, []string{LogFileName})
	LogTailerProcessQueueSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "log_tailer_processing_queue_size",
		Help: "log tailer processing queue size",
	}, []string{LogFileName})
)

func PromHandler(_ http.Handler) http.Handler {
	return promhttp.HandlerFor(defaultGatherer, promhttp.HandlerOpts{})
}
