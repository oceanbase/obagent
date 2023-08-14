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

package message

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Type string

const (
	Counter   Type = "Counter"
	Gauge     Type = "Gauge"
	Summary   Type = "Summary"
	Histogram Type = "Histogram"
	Untyped   Type = "Untyped"
	Log       Type = "log"
	Trace     Type = "trace"
)

type FieldEntry struct {
	Name  string
	Value interface{}
}

type TagEntry struct {
	Name  string
	Value string
}

// Message is the basic unit of the monitor
type Message struct {
	name        string
	fields      []FieldEntry
	tags        []TagEntry
	timestamp   time.Time
	msgType     Type
	tagSorted   bool
	fieldSorted bool
	id          string
}

func NewMessage(name string, msgType Type, timestamp time.Time) *Message {
	return &Message{
		name:      name,
		msgType:   msgType,
		timestamp: timestamp,
		tags:      make([]TagEntry, 0, 4),
		fields:    make([]FieldEntry, 0, 2),
	}
}

func NewMessageWithTagsFields(name string, msgType Type, timestamp time.Time, tags []TagEntry, fields []FieldEntry) *Message {
	return &Message{
		name:      name,
		msgType:   msgType,
		timestamp: timestamp,
		tags:      tags,
		fields:    fields,
	}
}

func (m *Message) Clone() *Message {
	fields := make([]FieldEntry, len(m.fields))
	tags := make([]TagEntry, len(m.tags))
	copy(fields, m.fields)
	copy(tags, m.tags)
	return &Message{
		name:      m.name,
		msgType:   m.msgType,
		timestamp: m.timestamp,
		tags:      tags,
		fields:    fields,
	}
}

func (m *Message) GetName() string {
	return m.name
}

func (m *Message) Rename(name string) {
	m.name = name
}

func (m *Message) GetTime() time.Time {
	return m.timestamp
}

func (m *Message) GetMetricType() Type {
	return m.msgType
}

func (m *Message) Fields() []FieldEntry {
	return m.fields
}

func (m *Message) AddField(name string, value interface{}) *Message {
	m.fields = append(m.fields, FieldEntry{Name: name, Value: value})
	m.fieldSorted = false
	return m
}

func (m *Message) SetField(name string, value interface{}) *Message {
	i, ok := m.findFieldEntry(name)
	if ok {
		m.fields[i].Value = value
	} else {
		m.AddField(name, value)
	}
	return m
}

func (m *Message) GetField(name string) (interface{}, bool) {
	i, ok := m.findFieldEntry(name)
	if !ok {
		return nil, false
	}
	return m.fields[i].Value, true
}

func (m *Message) RemoveField(name string) (interface{}, bool) {
	i, ok := m.findFieldEntry(name)
	if !ok {
		return nil, false
	}
	e := m.fields[i]
	ret := e.Value
	m.fields = append(m.fields[0:i], m.fields[i+1:]...)
	return ret, true
}

func (m *Message) RemoveAllFields() {
	m.fields = []FieldEntry{}
}

func (m *Message) findFieldEntry(name string) (int, bool) {
	if m.fieldSorted {
		i := sort.Search(len(m.fields), func(i int) bool {
			return m.fields[i].Name >= name
		})
		if i >= 0 && i < len(m.fields) && m.fields[i].Name == name {
			return i, true
		} else {
			return -1, false
		}
	}
	for i, e := range m.fields {
		if e.Name == name {
			return i, true
		}
	}
	return -1, false
}

func (m *Message) Tags() []TagEntry {
	return m.tags
}

func (m *Message) AddTag(name string, value string) *Message {
	m.tags = append(m.tags, TagEntry{Name: name, Value: value})
	m.tagSorted = false
	m.id = ""
	return m
}

func (m *Message) GetTag(name string) (string, bool) {
	i, ok := m.findTagEntry(name)
	if !ok {
		return "", false
	}
	return m.tags[i].Value, true
}

func (m *Message) GetAllTags(name string) []string {
	var ret []string
	for _, tag := range m.tags {
		if tag.Name == name {
			ret = append(ret, tag.Value)
		}
	}
	return ret
}

func (m *Message) RemoveTag(name string) (string, bool) {
	i, ok := m.findTagEntry(name)
	if !ok {
		return "", false
	}
	e := m.tags[i]
	ret := e.Value
	m.tags = append(m.tags[0:i], m.tags[i+1:]...)
	return ret, true
}

func (m *Message) findTagEntry(name string) (int, bool) {
	if m.tagSorted {
		i := sort.Search(len(m.tags), func(i int) bool {
			return m.tags[i].Name >= name
		})
		if i >= 0 && i < len(m.tags) && m.tags[i].Name == name {
			return i, true
		} else {
			return -1, false
		}
	}
	for i, e := range m.tags {
		if e.Name == name {
			return i, true
		}
	}
	return -1, false
}

func (m *Message) SetTag(name string, value string) *Message {
	i, ok := m.findTagEntry(name)
	if ok {
		m.tags[i].Value = value
	} else {
		m.AddTag(name, value)
	}
	return m
}

func (m *Message) SortTag() {
	if m.tagSorted {
		return
	}
	if len(m.tags) <= 1 {
		return
	}
	sort.Slice(m.tags, func(i, j int) bool {
		return m.tags[i].Name < m.tags[j].Name
	})
	m.tagSorted = true
}

func (m *Message) SortField() {
	if m.fieldSorted {
		return
	}
	if len(m.fields) <= 1 {
		return
	}
	sort.Slice(m.fields, func(i, j int) bool {
		return m.fields[i].Name < m.fields[j].Name
	})
	m.fieldSorted = true
}

func (m *Message) Identifier() string {
	if m.id != "" {
		return m.id
	}
	m.SortTag()
	size := len(m.name)
	for _, e := range m.tags {
		size += len(e.Name) + len(e.Value) + 2
	}
	sb := strings.Builder{}
	sb.Grow(size)
	sb.WriteString(m.name)
	for _, e := range m.tags {
		sb.WriteByte('\x00')
		sb.WriteString(e.Name)
		sb.WriteByte('\x00')
		sb.WriteString(e.Value)
	}
	m.id = sb.String()
	return m.id
}

func (m *Message) String() string {
	sb := strings.Builder{}
	sb.WriteString(m.name)
	sb.WriteByte('{')
	sb.WriteString(string(m.msgType))
	sb.WriteByte('@')
	sb.WriteString(m.timestamp.String())
	sb.WriteByte(':')
	for _, e := range m.tags {
		sb.WriteString(e.Name)
		sb.WriteByte('=')
		sb.WriteString(e.Value)
		sb.WriteByte(';')
	}
	sb.WriteByte('|')
	for _, e := range m.fields {
		sb.WriteString(e.Name)
		sb.WriteByte('=')
		sb.WriteString(fmt.Sprint(e.Value))
		sb.WriteByte(';')
	}
	sb.WriteByte('}')
	return sb.String()
}
