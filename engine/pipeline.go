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
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/metric"
	"github.com/oceanbase/obagent/stat"
)

//PipelineInstance pipeline working instance
type PipelineInstance struct {
	pipeline *pipeline
	name     string
	config   *config.PipelineConfig
	clock    *clock
}

type clock struct {
	context context.Context
	cancel  context.CancelFunc
	ticker  *time.Ticker
}

//Init initialize the pipeline instance.
//Including the initialization operation of the input processor output exporter.
//If err is nil, indicates successful initialization.
func (p *PipelineInstance) Init() error {
	err := p.pipeline.init()
	if err != nil {
		return errors.Wrap(err, "pipeline init")
	}
	return nil
}

//Start pipeline instance start working, divided into push and pull modes.
//Push mode, start the clock to push.
//Pull mode, register route with route manager and add pipeline.
func (p *PipelineInstance) Start() {
	switch p.config.ScheduleStrategy {
	case config.Trigger:
		p.registerRoute()
		GetRouteManager().addPipelineFromPipelineGroup(p.config.ExposeUrl, p)
	case config.Periodic:
		p.startClock(p.config.Period)
	}
}

//Stop pipeline instance stop working.
//Push mode, stop the clock to push.
//Pull mode, delete pipeline to the config manager.
func (p *PipelineInstance) Stop() {
	switch p.config.ScheduleStrategy {
	case config.Trigger:
		GetRouteManager().delPipelineFromPipelineGroup(p.config.ExposeUrl, p)
	case config.Periodic:
		p.clock.stopClock()
	}
}

func (p *PipelineInstance) registerRoute() {
	GetRouteManager().registerHTTPRoute(p.config.ExposeUrl)
}

func (p *PipelineInstance) parallelCompute() []metric.Metric {
	var waitGroup sync.WaitGroup
	var metricTotal []metric.Metric
	var metricMutex sync.Mutex
	for _, inputInstance := range p.pipeline.InputInstances {
		waitGroup.Add(1)
		go collectAndProcess(&waitGroup, &metricTotal, &metricMutex, inputInstance, p.pipeline.ProcessorInstances)
	}
	waitGroup.Wait()
	return metricTotal
}

func (p *PipelineInstance) startClock(duration time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(duration)
	c := &clock{
		context: ctx,
		cancel:  cancel,
		ticker:  ticker,
	}
	p.clock = c
	go func() {
		for {
			select {
			case <-c.ticker.C:
				p.pipelinePush()
			case <-c.context.Done():
				log.Info("pipeline clock is stop")
				return
			}
		}
	}()
}

func (c *clock) stopClock() {
	c.ticker.Stop()
	c.cancel()
}

//pipelinePush perform a pipeline push.
// ┌────────┐       ┌────────────┐       ┌────────────┐       ┌────────────┐
// │ Input1 │------>│ Processor1 │------>│ Processor2 │------>│ Processor3 │---┐
// └────────┘       └────────────┘       └────────────┘       └────────────┘   │
// ┌────────┐       ┌────────────┐       ┌────────────┐       ┌────────────┐   │    ┌────────┐
// │ Input2 │------>│ Processor1 │------>│ Processor2 │------>│ Processor3 │---┼--->│ Output │
// └────────┘       └────────────┘       └────────────┘       └────────────┘   │    └────────┘
// ┌────────┐       ┌────────────┐       ┌────────────┐       ┌────────────┐   │
// │ Input3 │------>│ Processor1 │------>│ Processor2 │------>│ Processor3 │---┘
// └────────┘       └────────────┘       └────────────┘       └────────────┘
func (p *PipelineInstance) pipelinePush() {
	tStart := time.Now()
	metrics := p.parallelCompute()
	if metrics == nil || len(metrics) == 0 {
		log.Warnf("push pipeline parallel compute result metrics is nil")
		return
	}
	err := p.pipeline.OutputInstance.Write(metrics)
	if err != nil {
		log.WithError(err).Error("pipeline output plugin write metric failed")
		return
	}
	elapsedTimeSeconds := time.Now().Sub(tStart).Seconds()
	stat.MonAgentPipelineExecuteTotal.With(prom.Labels{"name": p.name, "status": "Successful"}).Inc()
	stat.MonAgentPipelineExecuteSecondsTotal.With(prom.Labels{"name": p.name, "status": "Successful"}).Add(elapsedTimeSeconds)
}

func collectAndProcess(waitGroup *sync.WaitGroup, metricTotal *[]metric.Metric, metricMutex *sync.Mutex, inputInstance *InputInstance, processorInstance []*ProcessorInstance) {
	defer waitGroup.Done()

	metrics, err := inputInstance.Collect()
	if err != nil {
		log.WithError(err).Error("input plugin collect failed")
		return
	}
	for _, processorInstance := range processorInstance {
		metrics, err = processorInstance.Process(metrics...)
		if err != nil {
			log.WithError(err).Error("process plugin process failed")
			continue
		}
	}

	metricMutex.Lock()
	*metricTotal = append(*metricTotal, metrics...)
	metricMutex.Unlock()
}

type pipeline struct {
	InputInstances     []*InputInstance
	ProcessorInstances []*ProcessorInstance
	OutputInstance     *OutputInstance
	ExporterInstance   *ExporterInstance
}

func (p *pipeline) init() error {

	for _, inputInstance := range p.InputInstances {
		err := inputInstance.init(inputInstance.Config)
		if err != nil {
			return errors.Wrap(err, "input init")
		}
	}

	for _, processorInstance := range p.ProcessorInstances {
		err := processorInstance.init(processorInstance.Config)
		if err != nil {
			return errors.Wrap(err, "processor init")
		}
	}

	if p.OutputInstance != nil {
		err := p.OutputInstance.init(p.OutputInstance.Config)
		if err != nil {
			return errors.Wrap(err, "output init")
		}
	}

	if p.ExporterInstance != nil {
		err := p.ExporterInstance.init(p.ExporterInstance.Config)
		if err != nil {
			return errors.Wrap(err, "exporter init")
		}
	}

	return nil
}
