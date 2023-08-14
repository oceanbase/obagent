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

package retag

import (
	"context"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/monitor/message"
)

const sampleConfig = `
newTags:
  t1: v1
copyTags:
  t2: t1
renameTags:
  t3: t1
`

const description = `
modify metric tags, support add new tag, rename tag, copy a tag with new name
`

type RetagConfig struct {
	RenameTags map[string]string `yaml:"renameTags"`
	CopyTags   map[string]string `yaml:"copyTags"`
	NewTags    map[string]string `yaml:"newTags"`
}

type RetagProcessor struct {
	Config *RetagConfig
}

func (r *RetagProcessor) SampleConfig() string {
	return sampleConfig
}

func (r *RetagProcessor) Description() string {
	return description
}

func (r *RetagProcessor) Init(ctx context.Context, config map[string]interface{}) error {

	var retagConfig RetagConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "retagProcessor encode config")
	}

	err = yaml.Unmarshal(configBytes, &retagConfig)
	if err != nil {
		return errors.Wrap(err, "retagProcessor decode config")
	}

	r.Config = &retagConfig

	log.WithContext(ctx).Infof("init retagProcessor with config: %+v", r.Config)
	return nil
}

func (r *RetagProcessor) Start(in <-chan []*message.Message, out chan<- []*message.Message) (err error) {
	for messages := range in {
		outMessages, err := r.Process(context.Background(), messages...)
		if err != nil {
			log.Warnf("process message failed: %v", err)
		}
		out <- outMessages
	}
	return nil
}
func (r *RetagProcessor) Stop() {}

func (r *RetagProcessor) Process(ctx context.Context, metrics ...*message.Message) ([]*message.Message, error) {
	for _, metric := range metrics {
		if metric == nil {
			log.WithContext(ctx).Warnf("found nil message, skip process")
			continue
		}
		for k, v := range r.Config.NewTags {
			_, found := metric.GetTag(k)
			if !found {
				metric.AddTag(k, v)
			}
		}
		for k, newK := range r.Config.CopyTags {
			v, keyFound := metric.GetTag(k)
			if keyFound {
				_, newKeyFound := metric.GetTag(newK)
				if !newKeyFound {
					metric.AddTag(newK, v)
				}
			}
		}
		for k, newK := range r.Config.RenameTags {
			v, keyFound := metric.GetTag(k)
			if keyFound {
				_, newKeyFound := metric.GetTag(newK)
				if !newKeyFound {
					metric.AddTag(newK, v)
					metric.RemoveTag(k)
				}
			}
		}
	}
	log.WithContext(ctx).Debugf("after process, metrics length: %d", len(metrics))
	return metrics, nil
}
