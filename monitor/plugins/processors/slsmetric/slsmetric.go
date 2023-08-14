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

package slsmetric

import (
	"context"
	"sort"
	"strconv"

	"github.com/oceanbase/obagent/monitor/message"
)

type SlsMetricProcessor struct {
	ctx  context.Context
	done chan struct{}
}

func NewSlsMetricProcessor() *SlsMetricProcessor {
	return &SlsMetricProcessor{
		ctx:  context.Background(),
		done: make(chan struct{}),
	}
}

func (s *SlsMetricProcessor) Start(in <-chan []*message.Message, out chan<- []*message.Message) error {
	for msgs := range in {
		newMsgs := s.Process(s.ctx, msgs)
		out <- newMsgs
	}
	return nil
}

func (s *SlsMetricProcessor) Process(ctx context.Context, msgs []*message.Message) []*message.Message {
	var processedMsgs = make([]*message.Message, 0)
	fieldMsgs := message.ProcessFields(msgs)
	for _, msg := range fieldMsgs {
		tmpMsgs := processSingleMsg(msg)
		processedMsgs = append(processedMsgs, tmpMsgs...)
	}
	return processedMsgs
}

func processSingleMsg(msg *message.Message) []*message.Message {
	msgs := make([]*message.Message, 0)
	labelStr := labelsComb(msg.Tags())
	for _, field := range msg.Fields() {
		timeStr := strconv.FormatInt(msg.GetTime().UnixNano(), 10)
		tmpMsg := message.NewMessage(msg.GetName(), msg.GetMetricType(), msg.GetTime()).
			AddField("labels", labelStr).AddField("time_nano", timeStr).
			AddField(field.Name, field.Value)
		msgs = append(msgs, tmpMsg)
	}
	return msgs
}

func labelsComb(tags []message.TagEntry) string {
	sort.SliceStable(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	labelsStr := ""
	for i, k := range tags {
		labelsStr += k.Name
		labelsStr += "#$#"
		labelsStr += k.Value
		if i < len(tags)-1 {
			labelsStr += "|"
		}
	}
	return labelsStr
}

func (s *SlsMetricProcessor) Stop() {}
