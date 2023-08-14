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

package attr

import (
	"github.com/oceanbase/obagent/monitor/message"

	log "github.com/sirupsen/logrus"
)

type Oper string

const (
	addTagsOper      Oper = "addTags"
	copyTagsOper     Oper = "copyTags"
	renameTagsOper   Oper = "renameTags"
	removeTagsOper   Oper = "removeTags"
	removeFieldsOper Oper = "removeFields"
	removeMetricOper Oper = "removeMetric"
)

type Operation struct {
	Oper        Oper              `yaml:"oper"`
	Condition   Condition         `yaml:"condition"`
	Tags        map[string]string `yaml:"tags"`
	RemoveItems []string          `yaml:"removeItems"`
}

type Condition struct {
	Metric string
	Fields map[string]float64
	Tags   map[string]string
}

func (c *Condition) isMatched(metric *message.Message) bool {
	metricMatched := metric.GetName() == c.Metric || c.Metric == ""
	if !metricMatched {
		return false
	}
	fieldsMatched := true
	for name, value := range c.Fields {
		fieldVal, ex := metric.GetField(name)
		if !ex || fieldVal != value {
			fieldsMatched = false
			break
		}
	}
	if !fieldsMatched {
		return false
	}
	tagsMatched := true
	for name, value := range c.Tags {
		tagVal, ex := metric.GetTag(name)
		if !ex || tagVal != value {
			tagsMatched = false
			break
		}
	}
	if !tagsMatched {
		return false
	}

	return true
}

func switchOper(metric *message.Message, oper Operation) {
	if !oper.Condition.isMatched(metric) {
		return
	}
	switch oper.Oper {
	case addTagsOper:
		addTags(metric, oper.Tags)
		break
	case copyTagsOper:
		copyTags(metric, oper.Tags)
		break
	case renameTagsOper:
		renameTags(metric, oper.Tags)
		break
	case removeTagsOper:
		removeTags(metric, oper.RemoveItems)
		break
	case removeFieldsOper:
		removeFields(metric, oper.RemoveItems)
		break
	case removeMetricOper:
		metric.RemoveAllFields()
		break
	default:
		log.Errorf("oper %s is not supported.", oper.Oper)
	}
}

func addTags(metric *message.Message, tags map[string]string) {
	for k, v := range tags {
		metric.AddTag(k, v)
	}
}

func copyTags(metric *message.Message, tags map[string]string) {
	for oldKey, newKey := range tags {
		val, ex := metric.GetTag(oldKey)
		if !ex {
			continue
		}
		metric.AddTag(newKey, val)
	}
}

func renameTags(metric *message.Message, tags map[string]string) {
	for oldKey, newKey := range tags {
		val, ex := metric.GetTag(oldKey)
		if !ex {
			continue
		}
		metric.AddTag(newKey, val)
		metric.RemoveTag(oldKey)
	}
}

func removeTags(metric *message.Message, tags []string) {
	for _, tag := range tags {
		metric.RemoveTag(tag)
	}
}

func removeFields(metric *message.Message, fields []string) {
	for _, field := range fields {
		metric.RemoveField(field)
	}
}
