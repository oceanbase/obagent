package jointable

import (
	"github.com/oceanbase/obagent/monitor/message"
)

func operateMessage(cond Condition, msg *message.Message, multiTags []map[string]string) (removed bool) {
	// no dbData
	if len(multiTags) == 0 {
		return false
	}

	// metric or tags not matched, skip; not contains all tagNames, skip
	if !cond.isMetricMatched(msg) || !cond.isTagsMatched(msg) || !cond.containsAllTagNames(msg) {
		return false
	}

	// find the first matched message, then copy dbData tags to the message
	for _, tags := range multiTags {
		if cond.isTagNamesMatched(msg, tags) {
			// add tags
			copyTags(msg, tags)
			return false
		}
	}

	// containsAllTagNames but not any message matched, removed if RemoveNotMatchedTagValueMessage is true
	if cond.RemoveNotMatchedTagValueMessage {
		return true
	}
	return false
}

func copyTags(msg *message.Message, tags map[string]string) {
	for tag, tagValue := range tags {
		if _, ex := msg.GetTag(tag); !ex {
			msg.AddTag(tag, tagValue)
		}
	}
}

type Condition struct {
	Metrics []string        `yaml:"metrics"`
	metrics map[string]bool `yaml:"-"`
	// all tag key-value should be equal
	Tags map[string]string `yaml:"tags"`
	// all tag keys should be in tagNames
	TagNames []string `yaml:"tagNames"`
	// tags contain sonditions 's key, but not euqual all values
	RemoveNotMatchedTagValueMessage bool `yaml:"removeNotMatchedTagValueMessage"`
}

func (c *Condition) init() {
	c.metrics = make(map[string]bool)
	for _, metric := range c.Metrics {
		c.metrics[metric] = true
	}
}

func (c Condition) isMetricMatched(msg *message.Message) bool {
	return len(c.metrics) == 0 || c.metrics[msg.GetName()]
}

func (c Condition) isTagsMatched(msg *message.Message) bool {
	// all tags matched with msg.tags
	allTagsMatched := true
	for tag, value := range c.Tags {
		if tagVal, ex := msg.GetTag(tag); !ex || tagVal != value {
			allTagsMatched = false
			break
		}
	}
	return allTagsMatched
}

func (c Condition) containsAllTagNames(msg *message.Message) bool {
	for _, tag := range c.TagNames {
		if _, ex := msg.GetTag(tag); !ex {
			return false
		}
	}
	return true
}

func (c Condition) isTagNamesMatched(msg *message.Message, dbData map[string]string) bool {
	// if all tagNames in msg matched with dbData
	allTagNamesMatched := true
	for _, tag := range c.TagNames {
		dbDataVal, ex := dbData[tag]
		if !ex {
			continue
		}
		if tagVal, ex := msg.GetTag(tag); !ex || tagVal != dbDataVal {
			allTagNamesMatched = false
			break
		}
	}
	return allTagNamesMatched
}
