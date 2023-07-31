package attr

import (
	"context"
	"testing"
	"time"

	"github.com/oceanbase/obagent/monitor/message"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestAttr(t *testing.T) {
	configStr := `
operations:  
  - oper: addTags
    condition:
      tags:
        t1: tv1
    tags:
      f3: fv3
      f4: fv4
    removeItems:
    `
	var configMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &configMap)

	processor := &AttrProcessor{}
	processor.Init(context.Background(), configMap)

	in := make(chan []*message.Message, 2)
	out := make(chan []*message.Message, 2)
	go processor.Start(in, out)

	msg := message.NewMessage("test", message.Counter, time.Now())
	addTags(msg, map[string]string{
		"t1": "tv1",
		"t2": "tv2",
	})
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)
	in <- []*message.Message{msg}

	outmsg := <-out
	assert.Equal(t, 4, len(outmsg[0].Tags()))
}
