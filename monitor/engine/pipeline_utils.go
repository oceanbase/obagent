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
