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

package engine

import (
	"bytes"
	"context"
	"time"

	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/metric"
	"github.com/oceanbase/obagent/plugins"
	"github.com/oceanbase/obagent/stat"
)

//InputInstance input instance responsible for Collect work
type InputInstance struct {
	PluginName string
	Config     *config.PluginConfig
	Input      plugins.Input
}

//ProcessorInstance processor instance responsible for Process work
type ProcessorInstance struct {
	PluginName string
	Config     *config.PluginConfig
	Processor  plugins.Processor
}

//OutputInstance output instance responsible for Write work
type OutputInstance struct {
	PluginName string
	Config     *config.PluginConfig
	Output     plugins.Output
}

//ExporterInstance exporter instance responsible for Export work
type ExporterInstance struct {
	PluginName string
	Config     *config.PluginConfig
	Exporter   plugins.Exporter
}

func (i *InputInstance) init(config *config.PluginConfig) error {
	inputPlugin, err := plugins.GetInputManager().GetPlugin(i.PluginName)
	if err != nil {
		return errors.Wrapf(err, "input manager get plugin")
	}
	i.Input = inputPlugin
	err = inputPlugin.Init(config.PluginInnerConfig)
	if err != nil {
		return errors.Wrapf(err, "input plugin init")
	}
	return nil
}

func (p *ProcessorInstance) init(config *config.PluginConfig) error {
	processorPlugin, err := plugins.GetProcessorManager().GetPlugin(p.PluginName)
	if err != nil {
		return errors.Wrapf(err, "processor manager get plugin")
	}
	p.Processor = processorPlugin
	err = processorPlugin.Init(config.PluginInnerConfig)
	if err != nil {
		return errors.Wrapf(err, "processor plugin init")
	}
	return nil
}

func (o *OutputInstance) init(config *config.PluginConfig) error {
	outputPlugin, err := plugins.GetOutputManager().GetPlugin(o.PluginName)
	if err != nil {
		return errors.Wrapf(err, "output manager get plugin")
	}
	o.Output = outputPlugin
	err = outputPlugin.Init(config.PluginInnerConfig)
	if err != nil {
		return errors.Wrapf(err, "output plugin init")
	}
	return nil
}

func (e *ExporterInstance) init(config *config.PluginConfig) error {
	exporterPlugin, err := plugins.GetExporterManager().GetPlugin(e.PluginName)
	if err != nil {
		return errors.Wrapf(err, "exporter manager get plugin")
	}
	e.Exporter = exporterPlugin
	err = exporterPlugin.Init(config.PluginInnerConfig)
	if err != nil {
		return errors.Wrapf(err, "exporter plugin init")
	}
	return nil
}

type collectResult struct {
	metrics []metric.Metric
	err     error
}

//Collect return the collection results as metrics
func (i *InputInstance) Collect() ([]metric.Metric, error) {

	defer GoroutineProtection(log.WithField("pipeline", "input instance collect"))

	var metrics []metric.Metric
	var err error
	tStart := time.Now()
	if i.Config.Timeout != 0 {
		metrics, err = i.collectWithTimeout()
	} else {
		metrics, err = i.Input.Collect()
	}
	elapsedTimeSeconds := time.Now().Sub(tStart).Seconds()
	if err != nil {
		stat.MonAgentPluginExecuteTotal.With(prom.Labels{"name": i.PluginName, "status": "Error", "type": "Input"}).Inc()
		stat.MonAgentPluginExecuteSecondsTotal.With(prom.Labels{"name": i.PluginName, "status": "Error", "type": "Input"}).Add(elapsedTimeSeconds)
	} else {
		stat.MonAgentPluginExecuteTotal.With(prom.Labels{"name": i.PluginName, "status": "Successful", "type": "Input"}).Inc()
		stat.MonAgentPluginExecuteSecondsTotal.With(prom.Labels{"name": i.PluginName, "status": "Successful", "type": "Input"}).Add(elapsedTimeSeconds)
	}
	return metrics, err
}

func (i *InputInstance) collectWithTimeout() ([]metric.Metric, error) {
	var metrics []metric.Metric
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), i.Config.Timeout)
	defer cancel()

	done := make(chan *collectResult, 1)
	go func() {

		defer GoroutineProtection(log.WithField("pipeline", "input instance collect with timeout"))

		metrics, err := i.Input.Collect()
		done <- &collectResult{
			metrics: metrics,
			err:     err,
		}
	}()

	select {
	case r := <-done:
		metrics = r.metrics
		err = r.err
	case <-ctx.Done():
		err = errors.Errorf("input plugin %s collect time out", i.PluginName)
	}

	return metrics, err
}

type processorResult struct {
	metrics []metric.Metric
	err     error
}

//Process metrics content and return results
func (p *ProcessorInstance) Process(metrics ...metric.Metric) ([]metric.Metric, error) {

	defer GoroutineProtection(log.WithField("pipeline", "processor instance process"))

	var err error
	tStart := time.Now()
	if p.Config.Timeout != 0 {
		metrics, err = p.processWithTimeout(metrics...)
	} else {
		metrics, err = p.Processor.Process(metrics...)
	}
	elapsedTimeSeconds := time.Now().Sub(tStart).Seconds()
	if err != nil {
		stat.MonAgentPluginExecuteTotal.With(prom.Labels{"name": p.PluginName, "status": "Error", "type": "Processor"}).Inc()
		stat.MonAgentPluginExecuteSecondsTotal.With(prom.Labels{"name": p.PluginName, "status": "Error", "type": "Processor"}).Add(elapsedTimeSeconds)
	} else {
		stat.MonAgentPluginExecuteTotal.With(prom.Labels{"name": p.PluginName, "status": "Successful", "type": "Processor"}).Inc()
		stat.MonAgentPluginExecuteSecondsTotal.With(prom.Labels{"name": p.PluginName, "status": "Successful", "type": "Processor"}).Add(elapsedTimeSeconds)
	}
	return metrics, err
}

func (p *ProcessorInstance) processWithTimeout(metrics ...metric.Metric) ([]metric.Metric, error) {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), p.Config.Timeout)
	defer cancel()

	done := make(chan *processorResult, 1)
	go func() {

		defer GoroutineProtection(log.WithField("pipeline", "processor instance process with timeout"))

		metrics, err := p.Processor.Process(metrics...)
		done <- &processorResult{
			metrics: metrics,
			err:     err,
		}
	}()

	select {
	case r := <-done:
		metrics = r.metrics
		err = r.err
	case <-ctx.Done():
		err = errors.Errorf("processor plugin %s process time out", p.PluginName)
	}

	return metrics, err
}

type outputResult struct {
	err error
}

//Write metrics to output target
func (o *OutputInstance) Write(metrics []metric.Metric) error {

	defer GoroutineProtection(log.WithField("pipeline", "output instance write"))

	var err error
	tStart := time.Now()
	if o.Config.Timeout != 0 {
		err = o.writeWithTimeout(metrics...)
	} else {
		err = o.Output.Write(metrics)
	}
	elapsedTimeSeconds := time.Now().Sub(tStart).Seconds()
	if err != nil {
		stat.MonAgentPluginExecuteTotal.With(prom.Labels{"name": o.PluginName, "status": "Error", "type": "Output"}).Inc()
		stat.MonAgentPluginExecuteSecondsTotal.With(prom.Labels{"name": o.PluginName, "status": "Error", "type": "Output"}).Add(elapsedTimeSeconds)
	} else {
		stat.MonAgentPluginExecuteTotal.With(prom.Labels{"name": o.PluginName, "status": "Successful", "type": "Output"}).Inc()
		stat.MonAgentPluginExecuteSecondsTotal.With(prom.Labels{"name": o.PluginName, "status": "Successful", "type": "Output"}).Add(elapsedTimeSeconds)
		stat.MonAgentPipelineReportMetricsTotal.With(prom.Labels{"name": o.PluginName}).Add(float64(len(metrics)))
	}
	return err
}

func (o *OutputInstance) writeWithTimeout(metrics ...metric.Metric) error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), o.Config.Timeout)
	defer cancel()

	done := make(chan *outputResult, 1)
	go func() {

		defer GoroutineProtection(log.WithField("pipeline", "output instance write with timeout"))

		err := o.Output.Write(metrics)
		done <- &outputResult{
			err: err,
		}
	}()

	select {
	case r := <-done:
		err = r.err
	case <-ctx.Done():
		err = errors.Errorf("output plugin %s write time out", o.PluginName)
	}

	return err
}

type exporterResult struct {
	buffer *bytes.Buffer
	err    error
}

//Export metrics to buffers and return
func (e *ExporterInstance) Export(metrics []metric.Metric) (*bytes.Buffer, error) {

	defer GoroutineProtection(log.WithField("pipeline", "exporter instance export"))

	var buffer *bytes.Buffer
	var err error
	tStart := time.Now()
	if e.Config.Timeout != 0 {
		buffer, err = e.exportWithTimeout(metrics...)
	} else {
		buffer, err = e.Exporter.Export(metrics)
	}
	elapsedTimeSeconds := time.Now().Sub(tStart).Seconds()
	if err != nil {
		stat.MonAgentPluginExecuteTotal.With(prom.Labels{"name": e.PluginName, "status": "Error", "type": "Exporter"}).Inc()
		stat.MonAgentPluginExecuteSecondsTotal.With(prom.Labels{"name": e.PluginName, "status": "Error", "type": "Exporter"}).Add(elapsedTimeSeconds)
	} else {
		stat.MonAgentPluginExecuteTotal.With(prom.Labels{"name": e.PluginName, "status": "Successful", "type": "Exporter"}).Inc()
		stat.MonAgentPluginExecuteSecondsTotal.With(prom.Labels{"name": e.PluginName, "status": "Successful", "type": "Exporter"}).Add(elapsedTimeSeconds)
		stat.MonAgentPipelineReportMetricsTotal.With(prom.Labels{"name": e.PluginName}).Add(float64(len(metrics)))
	}
	return buffer, err
}

func (e *ExporterInstance) exportWithTimeout(metrics ...metric.Metric) (*bytes.Buffer, error) {
	var buffer *bytes.Buffer
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), e.Config.Timeout)
	defer cancel()

	done := make(chan *exporterResult, 1)
	go func() {

		defer GoroutineProtection(log.WithField("pipeline", "exporter instance export with timeout"))

		buffer, err := e.Exporter.Export(metrics)
		done <- &exporterResult{
			buffer: buffer,
			err:    err,
		}
	}()

	select {
	case r := <-done:
		buffer = r.buffer
		err = r.err
	case <-ctx.Done():
		err = errors.Errorf("exporter plugin %s export time out", e.PluginName)
	}

	return buffer, err
}
