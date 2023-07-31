package engine

import (
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/plugins"
)

func convertProcessorsToPipeFunc(processors []plugins.Processor) []plugins.PipeFunc {
	pipeFuncs := make([]plugins.PipeFunc, len(processors))
	for i, processor := range processors {
		pipeFuncs[i] = processor.Start
	}
	return pipeFuncs
}

func batchCloseChan(channels []chan []*message.Message) {
	for _, channel := range channels {
		close(channel)
	}
}
