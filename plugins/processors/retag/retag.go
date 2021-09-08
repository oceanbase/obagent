// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package retag

import (
	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/metric"
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

func (r *RetagProcessor) Close() error {
	return nil
}

func (r *RetagProcessor) SampleConfig() string {
	return sampleConfig
}

func (r *RetagProcessor) Description() string {
	return description
}

func (r *RetagProcessor) Init(config map[string]interface{}) error {

	var retagConfig RetagConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "retag processor encode config")
	}

	err = yaml.Unmarshal(configBytes, &retagConfig)
	if err != nil {
		return errors.Wrap(err, "retag processor decode config")
	}

	r.Config = &retagConfig

	log.Info("retag processor init with config", r.Config)
	return nil
}

func (r *RetagProcessor) Process(metrics ...metric.Metric) ([]metric.Metric, error) {
	for _, metric := range metrics {
		tags := metric.Tags()
		for k, v := range r.Config.NewTags {
			_, found := tags[k]
			if !found {
				tags[k] = v
			} else {
				log.Warnf("already exist tag %s, do not add new one", k)
			}
		}
		for k, v := range r.Config.CopyTags {
			_, keyFound := tags[k]
			if keyFound {
				_, newKeyFound := tags[v]
				if !newKeyFound {
					tags[v] = tags[k]
				} else {
					log.Warn("already exists new key:", v)
				}
			} else {
				log.Warn("key not found:", k)
			}
		}
		for k, v := range r.Config.RenameTags {
			_, keyFound := tags[k]
			if keyFound {
				_, newKeyFound := tags[v]
				if !newKeyFound {
					tags[v] = tags[k]
					delete(tags, k)
				} else {
					log.Warn("already exists new key:", v)
				}
			} else {
				log.Warn("key not found:", k)
			}
		}
	}
	log.Debug("after process:", metrics)
	return metrics, nil
}
