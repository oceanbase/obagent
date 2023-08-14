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

package aggregate

import (
	"context"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/utils"
)

const sampleConfig = `
rules:
  - metric: m1
    tags: [ t1, t2 ]
  - metric: m2
    tags: [ t1, t2 ]
`

const description = `
aggregate metrics
`

type Rule struct {
	MetricName string   `yaml:"metric"`
	Tags       []string `yaml:"tags"`
}

type AggregatorConfig struct {
	Rules                []*Rule `yaml:"rules"`
	IsRetainNativeMetric bool    `yaml:"isRetainNativeMetric"`
}

type Aggregator struct {
	Config *AggregatorConfig
	rules  map[string]*rule
}

func (a *Aggregator) SampleConfig() string {
	return sampleConfig
}

func (a *Aggregator) Description() string {
	return description
}

func (a *Aggregator) Init(ctx context.Context, config map[string]interface{}) error {

	var aggregatorConfig AggregatorConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "aggregateProcessor encode config")
	}

	err = yaml.Unmarshal(configBytes, &aggregatorConfig)
	if err != nil {
		return errors.Wrap(err, "aggregateProcessor decode config")
	}

	a.Config = &aggregatorConfig
	a.rules = make(map[string]*rule)
	for _, ruleConfig := range a.Config.Rules {
		a.rules[ruleConfig.MetricName] = ruleFromConfig(ruleConfig)
	}
	log.WithContext(ctx).Infof("init aggregateProcessor with config: %+v", a.Config)
	return nil
}

func (a *Aggregator) Start(in <-chan []*message.Message, out chan<- []*message.Message) error {
	for msgs := range in {
		newMsgs, err := a.Process(context.Background(), msgs...)
		if err != nil {
			log.Errorf("aggregator process messages failed, err %s", err)
		}
		out <- newMsgs
	}
	return nil
}

func (a *Aggregator) Stop() {}

type rule struct {
	key  string // message + group tags
	tags map[string]int
	f    func(a, b interface{}) interface{}
}

func ruleFromConfig(config *Rule) *rule {
	sort.Strings(config.Tags)
	m := make(map[string]int)
	for i, k := range config.Tags {
		m[k] = i
	}
	return &rule{
		key:  config.MetricName + "\x00" + strings.Join(config.Tags, "\x00"),
		tags: m,
		f:    sum,
	}
}

func (r *rule) groupKey(m *message.Message) (string, bool) {
	vs := make([]string, len(r.tags)+1)
	found := 0
	vs[0] = m.GetName()
	for _, e := range m.Tags() {
		i, ex := r.tags[e.Name]
		if ex {
			vs[i+1] = e.Value
			found++
		}
	}
	if found != len(vs)-1 {
		return "", false
	}
	return strings.Join(vs, "\x00"), true
}

func (r *rule) first(m *message.Message) *message.Message {
	tags := make([]message.TagEntry, 0, len(r.tags))
	for _, e := range m.Tags() {
		_, ex := r.tags[e.Name]
		if !ex {
			continue
		}
		tags = append(tags, e)
	}
	fields := make([]message.FieldEntry, len(m.Fields()))
	copy(fields, m.Fields())
	return message.NewMessageWithTagsFields(m.GetName(), m.GetMetricType(), m.GetTime(), tags, fields)
}

func sum(a, b interface{}) interface{} {
	fa, _ := utils.ConvertToFloat64(a)
	fb, _ := utils.ConvertToFloat64(b)
	return fa + fb
}

func (a *Aggregator) Process(ctx context.Context, metrics ...*message.Message) ([]*message.Message, error) {
	uniqueMetrics := message.UniqueMetrics(metrics) // todo: 去掉这里的 UniqueMetrics，把去重挪到对应 input
	ret := make([]*message.Message, 0, len(uniqueMetrics))
	aggrMap := make(map[string]*message.Message) // groupKey -> message
	for _, metricEntry := range uniqueMetrics {
		metricName := metricEntry.GetName()
		aRule, ok := a.rules[metricName]
		if !ok {
			ret = append(ret, metricEntry)
			continue
		}
		if a.Config.IsRetainNativeMetric {
			nativeMsg := metricEntry.Clone()
			nativeMsg.Rename(metricName + "_native")
			ret = append(ret, nativeMsg)
		}
		groupKey, ok := aRule.groupKey(metricEntry)
		if !ok {
			continue
		}
		msg, ok := aggrMap[groupKey]
		if !ok {
			msg = aRule.first(metricEntry)
			aggrMap[groupKey] = msg
		} else {
			for _, field := range metricEntry.Fields() {
				prev, _ := msg.GetField(field.Name)
				msg.SetField(field.Name, aRule.f(prev, field.Value))
			}
		}
	}
	for _, v := range aggrMap {
		ret = append(ret, v)
	}
	log.WithContext(ctx).Debugf("after process, metrics length: %d", len(ret))
	return ret, nil
}
