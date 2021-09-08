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

import "github.com/oceanbase/obagent/config"

func createInputInstance(pluginNode *config.PluginNode) *InputInstance {
	input := &InputInstance{
		PluginName: pluginNode.Plugin,
		Config:     pluginNode.Config,
		Input:      nil,
	}
	return input
}

func createProcessorInstance(pluginNode *config.PluginNode) *ProcessorInstance {
	processor := &ProcessorInstance{
		PluginName: pluginNode.Plugin,
		Config:     pluginNode.Config,
		Processor:  nil,
	}
	return processor
}

func createOutputInstance(pluginNode *config.PluginNode) *OutputInstance {
	output := &OutputInstance{
		PluginName: pluginNode.Plugin,
		Config:     pluginNode.Config,
		Output:     nil,
	}
	return output
}

func createExporterInstance(pluginNode *config.PluginNode) *ExporterInstance {
	exporter := &ExporterInstance{
		PluginName: pluginNode.Plugin,
		Config:     pluginNode.Config,
		Exporter:   nil,
	}
	return exporter
}

func createPipeline(pipelineStructure *config.PipelineStructure) *pipeline {
	inputs := make([]*InputInstance, len(pipelineStructure.Inputs))
	processors := make([]*ProcessorInstance, len(pipelineStructure.Processors))
	for idx, inputPluginNode := range pipelineStructure.Inputs {
		inputs[idx] = createInputInstance(inputPluginNode)
	}
	for idx, processorPluginNode := range pipelineStructure.Processors {
		processors[idx] = createProcessorInstance(processorPluginNode)
	}
	var output *OutputInstance
	if pipelineStructure.Output != nil {
		output = createOutputInstance(pipelineStructure.Output)
	}
	var exporter *ExporterInstance
	if pipelineStructure.Exporter != nil {
		exporter = createExporterInstance(pipelineStructure.Exporter)
	}
	pipeline := &pipeline{
		InputInstances:     inputs,
		ProcessorInstances: processors,
		OutputInstance:     output,
		ExporterInstance:   exporter,
	}
	return pipeline
}

func createPipelineInstance(pipelineNode *config.PipelineNode) *PipelineInstance {
	pipeline := createPipeline(pipelineNode.Structure)
	pipelineInstance := &PipelineInstance{
		name:     pipelineNode.Name,
		pipeline: pipeline,
		config:   pipelineNode.Config,
	}
	return pipelineInstance
}

//CreatePipelineInstances create pipeline instances according to the pipeline module
func CreatePipelineInstances(pipelineModule *config.PipelineModule) []*PipelineInstance {
	pipelines := make([]*PipelineInstance, len(pipelineModule.Pipelines))
	for idx, pipelineNode := range pipelineModule.Pipelines {
		pipelineInstance := createPipelineInstance(pipelineNode)
		pipelines[idx] = pipelineInstance
	}
	return pipelines
}
