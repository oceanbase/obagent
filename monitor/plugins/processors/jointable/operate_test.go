package jointable

import (
	"testing"
	"time"

	"github.com/huandu/go-assert"
	"github.com/oceanbase/obagent/monitor/message"
)

func TestIsMetricMatched(t *testing.T) {

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
				Metrics: []string{"not-matched-metric-name"},
			},
			matched: false,
		},
		{
			desc: "metric matched, tags and fields is empty",
			cond: Condition{
				Metrics: []string{"test"},
			},
			matched: true,
		},
	}

	msg := message.NewMessage("test", message.Counter, time.Now())
	msg.AddTag("t1", "tv1")
	msg.AddTag("t2", "tv2")
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)

	for _, testcase := range cases {
		testcase.cond.init()
		metricMatched := testcase.cond.isMetricMatched(msg)
		t.Log(testcase.desc)
		assert.Equal(t, testcase.matched, metricMatched)
	}
}

func TestIsTagsMatched(t *testing.T) {

	cases := []struct {
		desc    string
		cond    Condition
		matched bool
	}{
		{
			desc:    "tags is empty",
			cond:    Condition{},
			matched: true,
		},
		{
			desc: "tags is nil",
			cond: Condition{
				Tags: nil,
			},
			matched: true,
		},
		{
			desc: "some tags matched",
			cond: Condition{
				Tags: map[string]string{
					"t1": "tv1",
				},
			},
			matched: true,
		},
		{
			desc: "all tags matched",
			cond: Condition{
				Tags: map[string]string{
					"t1": "tv1",
					"t2": "tv2",
				},
			},
			matched: true,
		},
		{
			desc: "some tags not match",
			cond: Condition{
				Tags: map[string]string{
					"t1": "tv1",
					"t2": "not-matched",
				},
			},
			matched: false,
		},
		{
			desc: "any tags not match",
			cond: Condition{
				Tags: map[string]string{
					"t1": "not-matched",
					"t2": "not-matched",
				},
			},
			matched: false,
		},
	}

	msg := message.NewMessage("test", message.Counter, time.Now())
	msg.AddTag("t1", "tv1")
	msg.AddTag("t2", "tv2")
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)

	for _, testcase := range cases {
		testcase.cond.init()
		tagsMatched := testcase.cond.isTagsMatched(msg)
		t.Log(testcase.desc)
		assert.Equal(t, testcase.matched, tagsMatched)
	}
}

func TestContainsAllTagNames(t *testing.T) {

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
			desc: "condition only has some tag names",
			cond: Condition{
				TagNames: []string{"t1"},
			},
			matched: true,
		},
		{
			desc: "all tag names equal",
			cond: Condition{
				TagNames: []string{"t1", "t2"},
			},
			matched: true,
		},
		{
			desc: "all tag names not equal",
			cond: Condition{
				TagNames: []string{"x1", "x2"},
			},
			matched: false,
		},
		{
			desc: "some tag names not equal",
			cond: Condition{
				TagNames: []string{"t1", "x2"},
			},
			matched: false,
		},
	}

	msg := message.NewMessage("test", message.Counter, time.Now())
	msg.AddTag("t1", "tv1")
	msg.AddTag("t2", "tv2")
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)

	for _, testcase := range cases {
		testcase.cond.init()
		allTagNamesMatched := testcase.cond.containsAllTagNames(msg)
		t.Log(testcase.desc)
		assert.Equal(t, testcase.matched, allTagNamesMatched)
	}
}

func TestIsTagNamesMatched(t *testing.T) {
	cases := []struct {
		desc    string
		cond    Condition
		data    map[string]string
		matched bool
	}{
		{
			desc: "empty conditon",
			cond: Condition{},
			data: map[string]string{
				"t1": "tv1",
			},
			matched: true,
		},
		{
			desc: "tagNames 's value equal",
			cond: Condition{
				TagNames: []string{"t1"},
			},
			data: map[string]string{
				"t1": "tv1",
				"t2": "tv2",
			},
			matched: true,
		},
		{
			desc: "multi tagNames 's value equal",
			cond: Condition{
				TagNames: []string{"t1", "t2"},
			},
			data: map[string]string{
				"t1": "tv1",
				"t2": "tv2",
			},
			matched: true,
		},
		{
			desc: "tagNames not empty, but dbData absent",
			cond: Condition{
				TagNames: []string{"t1"},
			},
			data: map[string]string{
				"t2": "tv2",
			},
			matched: true,
		},
		{
			desc: "tagNames 's value not equal",
			cond: Condition{
				TagNames: []string{"t1"},
			},
			data: map[string]string{
				"t1": "xv1",
				"t2": "tv2",
			},
			matched: false,
		},
		{
			desc: "multi tagNames 's value not equal",
			cond: Condition{
				TagNames: []string{"t1", "t2"},
			},
			data: map[string]string{
				"t1": "xv1",
				"t2": "tv2",
			},
			matched: false,
		},
	}

	msg := message.NewMessage("test", message.Counter, time.Now())
	msg.AddTag("t1", "tv1")
	msg.AddTag("t2", "tv2")
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)

	for _, testcase := range cases {
		testcase.cond.init()
		isTagNamesMatched := testcase.cond.isTagNamesMatched(msg, testcase.data)
		t.Log(testcase.desc)
		assert.Equal(t, testcase.matched, isTagNamesMatched)
	}
}
