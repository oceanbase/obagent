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
	"testing"
	"time"

	"github.com/oceanbase/obagent/monitor/message"
	"github.com/stretchr/testify/assert"
)

func TestAddTags(t *testing.T) {
	cases := []struct {
		addTags    map[string]string
		wantedTags map[string]string
	}{
		{
			addTags: map[string]string{
				"t1": "tv1",
				"t2": "tv2",
			},
			wantedTags: map[string]string{
				"t1": "tv1",
				"t2": "tv2",
			},
		},
		{
			addTags:    map[string]string{},
			wantedTags: map[string]string{},
		},
	}
	for _, testcase := range cases {
		msg := message.NewMessage("test", message.Counter, time.Now())
		addTags(msg, testcase.addTags)
		assert.Equal(t, len(testcase.wantedTags), len(msg.Tags()))
		for _, tag := range msg.Tags() {
			val, ex := testcase.wantedTags[tag.Name]
			assert.Equal(t, true, ex)
			assert.Equal(t, val, tag.Value)
		}
	}
}

func TestCopyTags(t *testing.T) {
	cases := []struct {
		copyTags   map[string]string
		wantedTags map[string]string
	}{
		{
			copyTags: map[string]string{
				"t1": "new_t1",
				"t2": "new_t2",
			},
			wantedTags: map[string]string{
				"t1":     "tv1",
				"new_t1": "tv1",
			},
		},
		{
			copyTags: map[string]string{},
			wantedTags: map[string]string{
				"t1": "tv1",
			},
		},
	}
	for _, testcase := range cases {
		msg := message.NewMessage("test", message.Counter, time.Now())
		addTags(msg, map[string]string{
			"t1": "tv1",
		})
		copyTags(msg, testcase.copyTags)
		assert.Equal(t, len(testcase.wantedTags), len(msg.Tags()))
		for _, tag := range msg.Tags() {
			val, ex := testcase.wantedTags[tag.Name]
			assert.Equal(t, true, ex)
			assert.Equal(t, val, tag.Value)
		}
	}
}

func TestRenameTags(t *testing.T) {
	cases := []struct {
		renameTags map[string]string
		wantedTags map[string]string
	}{
		{
			renameTags: map[string]string{
				"t1": "new_t1",
				"t2": "new_t2",
			},
			wantedTags: map[string]string{
				"new_t1": "tv1",
			},
		},
		{
			renameTags: map[string]string{},
			wantedTags: map[string]string{
				"t1": "tv1",
			},
		},
	}
	for _, testcase := range cases {
		msg := message.NewMessage("test", message.Counter, time.Now())
		addTags(msg, map[string]string{
			"t1": "tv1",
		})

		renameTags(msg, testcase.renameTags)
		assert.Equal(t, len(testcase.wantedTags), len(msg.Tags()))
		for _, tag := range msg.Tags() {
			val, ex := testcase.wantedTags[tag.Name]
			assert.Equal(t, true, ex)
			assert.Equal(t, val, tag.Value)
		}
	}
}

func TestRemoveTags(t *testing.T) {
	cases := []struct {
		removeTags []string
		wantedTags map[string]string
	}{
		{
			removeTags: []string{"t1", "t2"},
			wantedTags: map[string]string{},
		},
		{
			removeTags: []string{"t1"},
			wantedTags: map[string]string{
				"t2": "tv2",
			},
		},
		{
			removeTags: []string{"no-exist-tag"},
			wantedTags: map[string]string{
				"t1": "tv1",
				"t2": "tv2",
			},
		},
		{
			removeTags: []string{},
			wantedTags: map[string]string{
				"t1": "tv1",
				"t2": "tv2",
			},
		},
	}
	for _, testcase := range cases {
		msg := message.NewMessage("test", message.Counter, time.Now())
		addTags(msg, map[string]string{
			"t1": "tv1",
			"t2": "tv2",
		})

		removeTags(msg, testcase.removeTags)
		for _, tag := range msg.Tags() {
			val, ex := testcase.wantedTags[tag.Name]
			assert.Equal(t, true, ex)
			assert.Equal(t, val, tag.Value)
		}
	}
}

func TestRemoveFields(t *testing.T) {
	cases := []struct {
		removeFields []string
		wantedFields map[string]interface{}
	}{
		{
			removeFields: []string{"f1", "f2"},
			wantedFields: map[string]interface{}{},
		},
		{
			removeFields: []string{"f1"},
			wantedFields: map[string]interface{}{
				"f2": 2.0,
			},
		},
		{
			removeFields: []string{"no-exist-field"},
			wantedFields: map[string]interface{}{
				"f1": 1.0,
				"f2": 2.0,
			},
		},
		{
			removeFields: []string{},
			wantedFields: map[string]interface{}{
				"f1": 1.0,
				"f2": 2.0,
			},
		},
	}
	for _, testcase := range cases {
		msg := message.NewMessage("test", message.Counter, time.Now())
		addTags(msg, map[string]string{
			"t1": "tv1",
			"t2": "tv2",
		})
		msg.AddField("f1", 1.0)
		msg.AddField("f2", 2.0)

		removeFields(msg, testcase.removeFields)
		for _, field := range msg.Fields() {
			val, ex := testcase.wantedFields[field.Name]
			assert.Equal(t, true, ex)
			assert.Equal(t, val, field.Value)
		}
	}
}

func TestRemoveMetric(t *testing.T) {
	oper := Operation{
		Condition: Condition{},
		Oper:      removeMetricOper,
	}
	msg := message.NewMessage("test", message.Counter, time.Now())
	addTags(msg, map[string]string{
		"t1": "tv1",
		"t2": "tv2",
	})
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)
	switchOper(msg, oper)

	assert.Equal(t, 0, len(msg.Fields()))
}

func TestIsMatch(t *testing.T) {

	cases := []struct {
		desc    string
		cond    Condition
		matched bool
	}{
		{
			desc:    "empty conditon",
			cond:    Condition{},
			matched: true,
		},
		{
			desc: "metric not matched",
			cond: Condition{
				Metric: "not-matched-metric-name",
			},
			matched: false,
		},
		{
			desc: "metric matched, tags and fields is empty",
			cond: Condition{
				Metric: "test",
			},
			matched: true,
		},
		{
			desc: "metric matched, tags and fields is nil",
			cond: Condition{
				Metric: "test",
				Tags:   nil,
				Fields: nil,
			},
			matched: true,
		},
		{
			desc: "metric matched, some tags matched, fields is nil",
			cond: Condition{
				Metric: "test",
				Tags: map[string]string{
					"t1": "tv1",
				},
				Fields: nil,
			},
			matched: true,
		},
		{
			desc: "metric matched, all tags matched, fields is nil",
			cond: Condition{
				Metric: "test",
				Tags: map[string]string{
					"t1": "tv1",
					"t2": "tv2",
				},
				Fields: nil,
			},
			matched: true,
		},
		{
			desc: "metric matched, some tags not match, fields is nil",
			cond: Condition{
				Metric: "test",
				Tags: map[string]string{
					"t1": "tv1",
					"t2": "not-matched",
				},
				Fields: nil,
			},
			matched: false,
		},
		{
			desc: "metric matched, any tags not match, fields is nil",
			cond: Condition{
				Metric: "test",
				Tags: map[string]string{
					"t1": "not-matched",
					"t2": "not-matched",
				},
				Fields: nil,
			},
			matched: false,
		},

		{
			desc: "metric matched, some fields matched, tags is nil",
			cond: Condition{
				Metric: "test",
				Tags:   nil,
				Fields: map[string]float64{
					"f1": 1.0,
				},
			},
			matched: true,
		},
		{
			desc: "metric matched, all fields matched, tags is nil",
			cond: Condition{
				Metric: "test",
				Tags:   nil,
				Fields: map[string]float64{
					"f1": 1.0,
					"f2": 2.0,
				},
			},
			matched: true,
		},
		{
			desc: "metric matched, some fields not match, tags is nil",
			cond: Condition{
				Metric: "test",
				Tags:   nil,
				Fields: map[string]float64{
					"f1": 1.0,
					"f2": 2.9999999,
				},
			},
			matched: false,
		},
		{
			desc: "metric matched, any fields not match, tags is nil",
			cond: Condition{
				Metric: "test",
				Tags:   nil,
				Fields: map[string]float64{
					"f1": 1.9999999,
					"f2": 2.9999999,
				},
			},
			matched: false,
		},
	}
	msg := message.NewMessage("test", message.Counter, time.Now())
	addTags(msg, map[string]string{
		"t1": "tv1",
		"t2": "tv2",
	})
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)

	for _, testcase := range cases {
		matched := testcase.cond.isMatched(msg)
		t.Log(testcase.desc)
		assert.Equal(t, testcase.matched, matched)
	}
}
