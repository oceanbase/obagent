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

package stat

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	//HTTPRequestTotal http request total
	_ = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_total",
		Help: "The total number of http request",
	},
		[]string{"method", "status", "path"})

	//HTTPRequestSecondsTotal gtto request seconds total
	_ = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_seconds_total",
		Help: "The total elapsed time in second of http request",
	},
		[]string{"method", "status", "path"})

	//MonAgentPipelineBufferMetrics monitor pipeline buffer metrics
	_ = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monagent_pipeline_buffer_metrics",
		Help: "The number of metrics currently in pipeline output buffer",
	},
		[]string{"name"})

	//MonAgentPipelineReportMetricsTotal monitor pipeline report metric total
	MonAgentPipelineReportMetricsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_pipeline_report_metrics_total",
		Help: "The total number of metrics reported from the pipeline",
	},
		[]string{"name"})

	//MonAgentPipelineExecuteTotal monitor pipeline execute total
	MonAgentPipelineExecuteTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_pipeline_execute_total",
		Help: "The total number of a pipeline executed",
	},
		[]string{"name", "status"})

	//MonAgentPipelineExecuteSecondsTotal monitor pipeline execute seconds total
	MonAgentPipelineExecuteSecondsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_pipeline_execute_seconds_total",
		Help: "The total time in second of a pipeline executed",
	},
		[]string{"name", "status"})

	//MonAgentPluginExecuteTotal monitor plugin execute total
	MonAgentPluginExecuteTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_plugin_execute_total",
		Help: "The total number of a plugin executed",
	},
		[]string{"name", "status", "type"})

	//MonAgentPluginExecuteSecondsTotal monitor plugin execute seconds total
	MonAgentPluginExecuteSecondsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monagent_plugin_execute_seconds_total",
		Help: "The total time in second of a plugin executed",
	},
		[]string{"name", "status", "type"})
)

func PromGinWrapper(_ http.Handler) http.Handler {
	return promhttp.Handler()
}
