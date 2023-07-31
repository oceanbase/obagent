package attr

import (
	"context"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/monitor/message"
)

const sampleConfig = `
operations:  
  - oper: addTags
    condition:
      tags:
        t1: tv1
    tags:
      f3: fv3
      f4: fv4
    removeItems:
`

const description = `message attr processor, such as addTags, copyTags, renameTags, removeTags, removeFields, removeMetric. Conditions must match the message attribitions. The removeItems is the tags or fields to be removed, if oper is removeMetric, removeItems can be empty.`

type AttrProcessorConfig struct {
	Operations []Operation `yaml:"operations"`
}

type AttrProcessor struct {
	Config AttrProcessorConfig
}

func (r *AttrProcessor) SampleConfig() string {
	return sampleConfig
}

func (r *AttrProcessor) Description() string {
	return description
}

func (r *AttrProcessor) Init(ctx context.Context, config map[string]interface{}) error {
	var attrConfig AttrProcessorConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "retagProcessor encode config")
	}

	err = yaml.Unmarshal(configBytes, &attrConfig)
	if err != nil {
		return errors.Wrap(err, "retagProcessor decode config")
	}

	r.Config = attrConfig

	log.WithContext(ctx).Infof("init retagProcessor with config: %+v", r.Config)
	return nil
}

func (r *AttrProcessor) Start(in <-chan []*message.Message, out chan<- []*message.Message) (err error) {
	for messages := range in {
		outMessages, err := r.Process(context.Background(), messages...)
		if err != nil {
			log.Warnf("process message failed: %v", err)
		}
		out <- outMessages
	}
	return nil
}

func (r *AttrProcessor) Stop() {}

func (r *AttrProcessor) Process(ctx context.Context, metrics ...*message.Message) ([]*message.Message, error) {
	for _, metric := range metrics {
		if metric == nil {
			log.WithContext(ctx).Warnf("found nil message, skip process")
			continue
		}
		for _, oper := range r.Config.Operations {
			switchOper(metric, oper)
		}
	}
	log.WithContext(ctx).Debugf("after process, metrics length: %d", len(metrics))
	return metrics, nil
}
