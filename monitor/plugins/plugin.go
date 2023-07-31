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
