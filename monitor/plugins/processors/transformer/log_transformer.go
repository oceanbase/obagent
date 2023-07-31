package transformer

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/log_analyzer"
	"github.com/oceanbase/obagent/monitor/message"
)

type LogTransformer struct {
}

func (r *LogTransformer) Init(ctx context.Context, config map[string]interface{}) error {
	log.WithContext(ctx).Infof("init obLogTransformer")
	return nil
}

func (r *LogTransformer) Start(in <-chan []*message.Message, out chan<- []*message.Message) (err error) {
	for messages := range in {
		outMessages, err := r.Process(context.Background(), messages...)
		if err != nil {
			log.Warnf("process message failed: %v", err)
		}
		out <- outMessages
	}
	return nil
}
func (r *LogTransformer) Stop() {}

func (r *LogTransformer) Process(ctx context.Context, metrics ...*message.Message) ([]*message.Message, error) {
	processedMessages := make([]*message.Message, 0)
	for _, metric := range metrics {
		metric.AddTag("app", log_analyzer.GetLogType(metric.GetName()))

		fieldsTags := make([]string, 0, 4)
		toBeRemovedTags := make([]string, 0, 4)
		for _, tag := range metric.Tags() {
			switch tag.Name {
			case "app", "level":
				continue
			default:
				fieldsTags = append(fieldsTags, tag.Name+":"+tag.Value)
				toBeRemovedTags = append(toBeRemovedTags, tag.Name)
			}
		}
		for _, toBeRemovedTag := range toBeRemovedTags {
			metric.RemoveTag(toBeRemovedTag)
		}
		metric.AddField("tags", fieldsTags)

		processedMessages = append(processedMessages, metric)
	}
	return processedMessages, nil
}
