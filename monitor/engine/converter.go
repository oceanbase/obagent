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

package engine

import (
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
)

// CreatePipelines create pipeline instances according to the pipeline module
func CreatePipelines(pipelineModule *monagent.PipelineModule) ([]*Pipeline, error) {
	pipelines := make([]*Pipeline, 0)
	for _, pipelineNode := range pipelineModule.Pipelines {
		if pipelineNode.Config.ScheduleStrategy != monagent.BySource {
			continue
		}
		pipelineInstance := NewPipeline(pipelineNode.Name, pipelineNode.Config)

		// set Source / Processor / Sink
		inputNodes := pipelineNode.Structure.Inputs
		sources := make([]plugins.Source, len(inputNodes))
		for i, inputPluginNode := range inputNodes {
			source, err := plugins.GetInputManager().GetPlugin(inputPluginNode.Plugin, inputPluginNode.Config)
			if err != nil {
				log.WithError(err).Error("CreatePipelineV2s init source failed")
				return nil, err
			}
			sources[i] = source
		}

		processorNodes := pipelineNode.Structure.Processors
		processors := make([]plugins.Processor, len(processorNodes))
		for i, processorPluginNode := range processorNodes {
			processor, err := plugins.GetProcessorManager().GetPlugin(processorPluginNode.Plugin, processorPluginNode.Config)
			if err != nil {
				log.WithError(err).Error("CreatePipelineV2s init processor failed")
				return nil, err
			}
			processors[i] = processor
		}

		var sink plugins.Sink
		outputNode := pipelineNode.Structure.Output
		if outputNode != nil {
			var err error
			sink, err = plugins.GetOutputManager().GetPlugin(outputNode.Plugin, outputNode.Config)
			if err != nil {
				log.WithError(err).Error("CreatePipelineV2s init sink failed")
				return nil, err
			}
		}

		exporterNode := pipelineNode.Structure.Exporter
		if exporterNode != nil {
			var err error
			sink, err = plugins.GetExporterManager().GetPlugin(exporterNode.Plugin, exporterNode.Config)
			if err != nil {
				log.WithError(err).Error("CreatePipelineV2s init sink failed")
				return nil, err
			}
		}

		pipelineInstance.SetSource(sources).SetProcessor(processors).SetSink(sink)
		pipelines = append(pipelines, pipelineInstance)
	}
	return pipelines, nil
}
