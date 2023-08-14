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
	"context"
	"github.com/oceanbase/obagent/monitor/plugins"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/monitor/message"
)

type Pipeline struct {
	Name           string
	Config         *monagent.PipelineConfig
	sources        []plugins.Source
	processors     []plugins.Processor
	sink           plugins.Sink
	sourceChannels []chan []*message.Message
}

const chanBufferSize = 500

func NewPipeline(name string, conf *monagent.PipelineConfig) *Pipeline {
	return &Pipeline{
		Name:   name,
		Config: conf,
	}
}

func (p *Pipeline) SetSource(sources []plugins.Source) *Pipeline {
	p.sources = sources
	return p
}

func (p *Pipeline) SetProcessor(processors []plugins.Processor) *Pipeline {
	p.processors = processors
	return p
}

func (p *Pipeline) SetSink(sink plugins.Sink) *Pipeline {
	p.sink = sink
	return p
}

func (p *Pipeline) Start(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			log.WithError(err).Errorf("start pipeline failed, pipeline: %s", p.Name)
			p.Stop()
		}
	}()
	ctxLog := log.WithContext(ctx)
	if len(p.sources) == 0 || p.sink == nil {
		err = errors.New("invalid sources / sink")
		ctxLog.WithFields(log.Fields{
			"sources": p.sources,
			"sink":    p.sink,
		}).WithError(err).Error()
		return
	}
	p.sourceChannels = make([]chan []*message.Message, len(p.sources))
	for i, source := range p.sources {
		p.sourceChannels[i] = make(chan []*message.Message)
		go func(src plugins.Source, c chan []*message.Message) {
			err1 := src.Start(c)
			if err1 != nil {
				ctxLog.WithError(err1).Error("start source failed")
			}
		}(source, p.sourceChannels[i])
	}

	convergedChan := p.converge(p.sourceChannels)
	sinkChannel, err := p.serialize(convergedChan, convertProcessorsToPipeFunc(p.processors))
	if err != nil {
		ctxLog.WithError(err).Error("executing processors failed")
		return err
	}

	go func() {
		err = p.sink.Start(sinkChannel)
		if err != nil {
			ctxLog.WithError(err).Error("pipeline start sink failed")
			return
		}
	}()

	return nil
}

func (p *Pipeline) Stop() {
	for _, source := range p.sources {
		source.Stop()
	}
	batchCloseChan(p.sourceChannels)
	for _, processor := range p.processors {
		processor.Stop()
	}
	p.sink.Stop()
}

// converge Data streams used to converge multiple sources
func (p *Pipeline) converge(inputs []chan []*message.Message) <-chan []*message.Message {
	wg := &sync.WaitGroup{}
	wg.Add(len(inputs))
	out := make(chan []*message.Message)

	for _, c := range inputs {
		go func(c <-chan []*message.Message) {
			defer wg.Done()
			for msgBatch := range c {
				out <- msgBatch
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// serialize Execute multiple Pipefuncs in serial
func (p *Pipeline) serialize(in <-chan []*message.Message, processors []plugins.PipeFunc) (<-chan []*message.Message, error) {
	if len(processors) == 0 {
		return in, nil
	}

	preOutChan := in
	for i := range processors {
		if i == len(processors)-1 {
			break
		}
		outChan := make(chan []*message.Message)
		go func(idx int, preChan <-chan []*message.Message, curChan chan []*message.Message) {
			err := processors[idx](preChan, curChan)
			if err != nil {
			}
			close(curChan)
		}(i, preOutChan, outChan)

		preOutChan = outChan
	}
	out := make(chan []*message.Message, 500)

	go func(preChan <-chan []*message.Message, curChan chan []*message.Message) {
		err := processors[len(processors)-1](preChan, curChan)
		if err != nil {
		}
		close(curChan)
	}(preOutChan, out)

	return out, nil
}
