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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestMetric(t time.Time) *Message {
	name := "test"
	metricType := Gauge
	metric := NewMessage(name, metricType, t).
		AddTag("k1", "v1").
		AddTag("k2", "v2").
		AddField("f1", 1.0).
		AddField("f2", 2.0)
	return metric
}

func TestNewMetric(t *testing.T) {
	ts := time.Now()
	metric := newTestMetric(ts)

	require.Equal(t, "test", metric.GetName())
	require.Equal(t, Gauge, metric.GetMetricType())

	s, ok := metric.GetTag("k1")
	require.True(t, ok)
	require.Equal(t, "v1", s)

	s, ok = metric.GetTag("k2")
	require.True(t, ok)
	require.Equal(t, "v2", s)

	v, ok := metric.GetField("f1")
	require.True(t, ok)
	require.Equal(t, 1.0, v)

	v, ok = metric.GetField("f2")
	require.True(t, ok)
	require.Equal(t, 2.0, v)

	require.Equal(t, ts, metric.GetTime())

	msg := NewMessageWithTagsFields("name", Log, time.Now(), []TagEntry{}, []FieldEntry{})
	require.Equal(t, "name", msg.GetName())
}

func TestCloneMetric(t *testing.T) {
	m := newTestMetric(time.Now())
	mc := m.Clone()
	require.True(t, reflect.DeepEqual(m, mc))
}

func TestMetricIdentity(t *testing.T) {
	m := newTestMetric(time.Now())
	id := m.Identifier()
	require.Equal(t, "test\x00k1\x00v1\x00k2\x00v2", id)
}

func TestTag(t *testing.T) {
	m := newTestMetric(time.Now())
	m.SetTag("k1", "V1")
	v, ok := m.GetTag("k1")
	require.True(t, ok)
	require.Equal(t, "V1", v)
	v, ok = m.RemoveTag("k2")
	require.True(t, ok)
	require.Equal(t, "v2", v)
	v, ok = m.GetTag("k2")
	require.False(t, ok)
}

func TestTagSort(t *testing.T) {
	m := newTestMetric(time.Now())
	m.AddTag("k5", "v5").AddTag("k4", "v4").AddTag("k3", "v3")
	fmt.Println(m.String())
	m.SortTag()
	fmt.Println(m.String())
	require.True(t, m.tagSorted)

	v, ok := m.GetTag("k1")
	require.True(t, ok)
	require.Equal(t, "v1", v)
	v, ok = m.GetTag("k3")
	require.True(t, ok)
	require.Equal(t, "v3", v)

	v, ok = m.RemoveTag("k2")
	require.True(t, ok)
	require.Equal(t, "v2", v)
	require.True(t, m.tagSorted)

	v, ok = m.GetTag("k2")
	require.False(t, ok)

	v, ok = m.GetTag("k9")
	require.False(t, ok)

	m.AddTag("k6", "v6")
	require.False(t, m.tagSorted)
}

func TestField(t *testing.T) {
	m := newTestMetric(time.Now())
	m.SetField("f1", 100.0)
	v, ok := m.GetField("f1")
	require.True(t, ok)
	require.Equal(t, 100.0, v)
	m.SetField("f3", 3.0)
	v, ok = m.GetField("f3")
	require.True(t, ok)
	require.Equal(t, 3.0, v)

	m.SortField()
	v, ok = m.GetField("f1")
	require.True(t, ok)
	require.Equal(t, 100.0, v)

}
