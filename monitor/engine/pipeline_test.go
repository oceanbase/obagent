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

package engine

import (
	"context"
	"fmt"
	"github.com/oceanbase/obagent/monitor/plugins"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/monitor/message"
)

func TestPipeline_converge(t *testing.T) {
	inputChans := make([]chan []*message.Message, 10)
	for i := range inputChans {
		inputChans[i] = make(chan []*message.Message)
	}

	type args struct {
		inputs []chan []*message.Message
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "汇聚多个 channel 数据流到一个 channel",
			args: args{
				inputs: inputChans,
			},
		},
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			p := &Pipeline{}
			go func() {
				for i, input := range tt.args.inputs {
					input <- []*message.Message{
						message.NewMessage(fmt.Sprintf("%d", i), message.Gauge, time.Now()),
					}
				}
				batchCloseChan(inputChans)
			}()
			count := 0
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				outChan := p.converge(inputChans)
				defer wg.Done()
				for messages := range outChan {
					count += len(messages)
				}
			}()
			wg.Wait()
			So(len(inputChans), ShouldEqual, count)
		})
	}
}

func TestPipeline_serialize(t *testing.T) {
	inChan := make(chan []*message.Message)
	var p1 plugins.PipeFunc = func(in <-chan []*message.Message, out chan<- []*message.Message) (err error) {
		for messages := range in {
			if len(messages) == 2 {
				messages[0].SetTag("testKey", "testValue")
				out <- messages
			}
		}
		return nil
	}

	var p2 plugins.PipeFunc = func(in <-chan []*message.Message, out chan<- []*message.Message) (err error) {
		for messages := range in {
			if len(messages) == 2 {
				messages[1].SetTag("testKey1", "testValue1")
				out <- messages
			}
		}
		return nil
	}
	processors := []plugins.PipeFunc{p1, p2}
	type args struct {
		in         chan []*message.Message
		processors []plugins.PipeFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "串行处理输入",
			args: args{
				in:         inChan,
				processors: processors,
			},
		},
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			p := &Pipeline{}
			go func() {
				tt.args.in <- []*message.Message{
					message.NewMessage("test", message.Gauge, time.Now()),
					message.NewMessage("test1", message.Gauge, time.Now()),
				}
				close(tt.args.in)
			}()
			outChan, err := p.serialize(tt.args.in, tt.args.processors)
			So(err, ShouldBeNil)
			for messages := range outChan {
				So(len(messages), ShouldEqual, 2)
				if len(messages) == 2 {
					val, ok := messages[0].GetTag("testKey")
					So(ok, ShouldBeTrue)
					So(val, ShouldEqual, "testValue")
					val1, ok1 := messages[1].GetTag("testKey1")
					So(ok1, ShouldBeTrue)
					So(val1, ShouldEqual, "testValue1")
				}
			}
		})
	}
}

type mockSource1 struct{}

func (s *mockSource1) Start(out chan<- []*message.Message) error {
	out <- []*message.Message{
		message.NewMessage("test1_1", message.Gauge, time.Now()),
		message.NewMessage("test1_2", message.Gauge, time.Now()),
	}
	return nil
}
func (s *mockSource1) Stop() {}

type mockSource2 struct{}

func (s *mockSource2) Start(out chan<- []*message.Message) error {
	out <- []*message.Message{
		message.NewMessage("test2_1", message.Gauge, time.Now()),
		message.NewMessage("test2_2", message.Gauge, time.Now()),
	}
	return nil
}
func (s *mockSource2) Stop() {}

type mockProcessor1 struct{}

func (s *mockProcessor1) Start(in <-chan []*message.Message, out chan<- []*message.Message) error {
	for messages := range in {
		if len(messages) == 2 {
			messages[0].SetTag("testKey", "testValue")
			out <- messages
		}
	}
	return nil
}
func (s *mockProcessor1) Stop() {}

type mockProcessor2 struct{}

func (s *mockProcessor2) Start(in <-chan []*message.Message, out chan<- []*message.Message) error {
	for messages := range in {
		if len(messages) == 2 {
			messages[1].SetTag("testKey1", "testValue1")
			out <- messages
		}
	}
	return nil
}
func (s *mockProcessor2) Stop() {}

type mockSink struct {
	count int
}

func (s *mockSink) Start(in <-chan []*message.Message) error {
	for messages := range in {
		if len(messages) == 2 {
			s.count++
		}
	}
	return nil
}
func (s *mockSink) Stop() {}

func (s *mockSink) getCount() int {
	return s.count
}

func TestPipeline_Start(t *testing.T) {
	type args struct {
		ctx        context.Context
		sources    []plugins.Source
		processors []plugins.Processor
		sink       plugins.Sink
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "执行 pipeline start",
			args: args{
				ctx:        context.Background(),
				sources:    []plugins.Source{&mockSource1{}, &mockSource2{}},
				processors: []plugins.Processor{&mockProcessor1{}, &mockProcessor2{}},
				sink:       &mockSink{},
			},
		},
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			p := NewPipeline("test", nil).
				SetSource(tt.args.sources).
				SetProcessor(tt.args.processors).
				SetSink(tt.args.sink)
			go func() {
				p.Start(tt.args.ctx)
			}()

			p.Stop()
		})
	}
}
