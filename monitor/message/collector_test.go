package message

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func newMessages() []*Message {
	var msgs = make([]*Message, 0)
	entry1 := NewMessage("test_message", Gauge, time.Now()).
		AddTag("dev", "sda").
		AddField("value", 1.0)
	entry2 := NewMessage("test_message", Summary, time.Now()).
		AddTag("dev", "sdb").
		AddField("value", 2.0)
	entry3 := NewMessage("test_message", Histogram, time.Now()).
		AddTag("dev", "sdc").
		AddField("value", 3.0)
	msgs = append(msgs, entry1, entry2, entry3)
	return msgs
}

func TestRegistry(t *testing.T) {
	msgs := newMessages()
	mf := CreateMetricFamily(msgs)
	collector := NewCollector(nil)
	collector.Fam = mf
	registry := prometheus.NewRegistry()
	err := registry.Register(collector)
	require.Equal(t, nil, err)
}
