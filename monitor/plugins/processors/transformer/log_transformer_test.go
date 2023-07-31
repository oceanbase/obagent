package transformer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/monitor/message"
)

func TestLogTransformer_Process(t *testing.T) {

	logTransformer := &LogTransformer{}
	msg := message.NewMessage("test.log", message.Log, time.Now())
	msg.AddTag("level", "info")
	msg.AddTag("ip", "127.0.0.1")
	msg.AddTag("obClusterId", "1")
	msg.AddTag("obClusterName", "testCluster")
	processedMsgs, err := logTransformer.Process(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(processedMsgs))
	if len(processedMsgs) != 0 {
		processedMsg1 := processedMsgs[0]
		app, ok := processedMsg1.GetTag("app")
		assert.Equal(t, "test", app)
		assert.True(t, ok)
		level, ok := processedMsg1.GetTag("level")
		assert.Equal(t, "info", level)
		assert.True(t, ok)
		tags, ok := processedMsg1.GetField("tags")
		assert.True(t, ok)
		tagsArr, ok := tags.([]string)
		assert.True(t, ok)
		assert.Equal(t, 3, len(tagsArr))
	}
}
