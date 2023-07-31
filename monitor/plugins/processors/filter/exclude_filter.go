package filter

import (
	"context"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/monitor/message"
)

const sampleConfig = `
conditions:
  - metric: m1
    tags:
      t1: v1
      t2: v2
  - metric: m2
    tags:
      t1: v1
      t2: v2
`

const description = `
filter out metrics match condition
`

type MetricFilterCondition struct {
	MetricName string             `yaml:"metric"`
	Tags       map[string]string  `yaml:"tags"`
	Fields     map[string]float64 `yaml:"fields"`
}

type ExcludeFilterConfig struct {
	Conditions []*MetricFilterCondition `yaml:"conditions"`
}

type ExcludeFilter struct {
	Config *ExcludeFilterConfig
}

func (e *ExcludeFilter) SampleConfig() string {
	return sampleConfig
}

func (e *ExcludeFilter) Description() string {
	return description
}

func (e *ExcludeFilter) Init(ctx context.Context, config map[string]interface{}) error {

	var excludeFilterConfig ExcludeFilterConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "excludeFilterProcessor encode config")
	}

	err = yaml.Unmarshal(configBytes, &excludeFilterConfig)
	if err != nil {
		return errors.Wrap(err, "excludeFilterProcessor decode config")
	}

	e.Config = &excludeFilterConfig

	log.WithContext(ctx).Infof("init excludeFilterProcessor with config: %+v", e.Config)
	return nil
}

func (e *ExcludeFilter) Start(in <-chan []*message.Message, out chan<- []*message.Message) error {
	for msgs := range in {
		newMsgs, err := e.Process(context.Background(), msgs...)
		if err != nil {
			log.Errorf("excludeFilter process messages failed, err: %s", err)
		}
		out <- newMsgs
	}
	return nil
}

func (e *ExcludeFilter) Stop() {}

func (e *ExcludeFilter) Process(ctx context.Context, metrics ...*message.Message) ([]*message.Message, error) {
	reservedMetrics := make([]*message.Message, 0, len(metrics))
	for _, metricEntry := range metrics {
		name := metricEntry.GetName()
		match := false
		for _, condition := range e.Config.Conditions {
			if condition.MetricName == name {
				match = true
				for tagName, tagValue := range condition.Tags {
					v, exists := metricEntry.GetTag(tagName)
					if !(exists && v == tagValue) {
						match = false
						break
					}
				}

				// check fields matched
				fieldsMatch := true
				for name, value := range condition.Fields {
					v, exists := metricEntry.GetField(name)
					if exists && v != value {
						fieldsMatch = false
						break
					}
				}
				match = match && (len(condition.Fields) == 0 || fieldsMatch)
			}
			if match {
				break
			}
		}
		if !match {
			reservedMetrics = append(reservedMetrics, metricEntry)
		}
	}
	log.WithContext(ctx).Debugf("after process, metrics length: %d", len(reservedMetrics))
	return reservedMetrics, nil
}
