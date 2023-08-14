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

package plugins

import (
	"context"

	"github.com/oceanbase/obagent/monitor/message"
)

// Source Data sources, production data
type Source interface {
	Start(out chan<- []*message.Message) (err error)
	Stop()
}

type Input interface {
	Collect(ctx context.Context) []*message.Message
}

// Processor Process the data
type Processor interface {
	Start(in <-chan []*message.Message, out chan<- []*message.Message) (err error)
	Stop()
}

type Mapper interface {
	Map(in *message.Message) *message.Message
}

// Sink Output the data in a specified manner
type Sink interface {
	Start(in <-chan []*message.Message) error
	Stop()
}

type PipeFunc func(in <-chan []*message.Message, out chan<- []*message.Message) (err error)
