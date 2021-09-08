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

package metric

import (
	"time"
)

type Type string

const (
	Counter   Type = "Counter"
	Gauge     Type = "Gauge"
	Summary   Type = "Summary"
	Histogram Type = "Histogram"
	Untyped   Type = "Untyped"
)

//Metric metric is the basic unit of the monitor
type Metric interface {
	Clone() Metric
	SetName(name string)
	GetName() string
	SetTime(time time.Time)
	GetTime() time.Time
	SetMetricType(metricType Type)
	GetMetricType() Type
	Fields() map[string]interface{}
	Tags() map[string]string
}

type metric struct {
	name       string
	fields     map[string]interface{}
	tags       map[string]string
	timestamp  time.Time
	metricType Type
}

func NewMetric(name string, fields map[string]interface{}, tags map[string]string, timestamp time.Time, metricType Type) *metric {
	return &metric{name: name, fields: fields, tags: tags, timestamp: timestamp, metricType: metricType}
}

func (m *metric) Clone() Metric {
	fields := make(map[string]interface{}, len(m.Fields()))
	tags := make(map[string]string, len(m.Tags()))
	for k, v := range m.Fields() {
		fields[k] = v
	}
	for k, v := range m.Tags() {
		tags[k] = v
	}
	return NewMetric(m.GetName(), fields, tags, m.GetTime(), m.GetMetricType())
}

func (m *metric) SetName(name string) {
	m.name = name
}

func (m *metric) GetName() string {
	return m.name
}

func (m *metric) SetTime(time time.Time) {
	m.timestamp = time
}

func (m *metric) GetTime() time.Time {
	return m.timestamp
}

func (m *metric) SetMetricType(metricType Type) {
	m.metricType = metricType
}

func (m *metric) GetMetricType() Type {
	return m.metricType
}

func (m *metric) Fields() map[string]interface{} {
	return m.fields
}

func (m *metric) Tags() map[string]string {
	return m.tags
}
