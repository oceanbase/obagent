package net

import (
	"context"
	"net"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/lib/trace"
	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/monitor/message"
)

const sampleConfig = `
timeout: 10s
targets:
  t1: '1.1.1.1:8080'
  t2: '2.2.2.2:8080'
`

const description = `
check network connectivity
`

type ConnectivityConfig struct {
	Timeout         time.Duration     `yaml:"timeout"`
	Targets         map[string]string `yaml:"targets"`
	CollectInterval time.Duration     `yaml:"collect_interval"`
}

type ConnectivityInput struct {
	Config *ConnectivityConfig
	ctx    context.Context
	done   chan struct{}
}

func (c *ConnectivityInput) Init(ctx context.Context, config map[string]interface{}) error {
	var pluginConfig ConnectivityConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "connectivity input encode config")
	}
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "connectivity input decode config")
	}
	c.Config = &pluginConfig
	c.ctx = ctx
	c.done = make(chan struct{})

	return nil
}

func (c *ConnectivityInput) Start(out chan<- []*message.Message) error {
	log.WithContext(c.ctx).Info("connectivityInput started")
	go c.update(c.ctx, out)
	return nil
}

func (c *ConnectivityInput) update(ctx context.Context, out chan<- []*message.Message) {
	ticker := time.NewTicker(c.Config.CollectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			connectivityMsgs, err := c.CollectMsgs()
			if err != nil {
				log.WithContext(ctx).Warnf("collect connectivity messages failed, err: %s", err)
			}
			out <- connectivityMsgs
		case <-c.done:
			log.WithContext(ctx).Info("connectivityInput exited")
			return
		}
	}
}

func (c *ConnectivityInput) Stop() {
	if c.done != nil {
		close(c.done)
	}
}

func (c *ConnectivityInput) SampleConfig() string {
	return sampleConfig
}

func (c *ConnectivityInput) Description() string {
	return description
}

func (c *ConnectivityInput) CollectMsgs() ([]*message.Message, error) {
	ctx := trace.ContextWithRandomTraceId()
	metrics := make([]*message.Message, 0, len(c.Config.Targets))
	for target, address := range c.Config.Targets {
		entry := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, time.Now())).WithField("connect address", address)
		conn, err := net.DialTimeout("tcp", address, c.Config.Timeout)
		entry.Debug("connect end")
		value := 0.0
		if err == nil && conn != nil {
			value = 1.0
			conn.Close()
		} else {
			log.WithContext(ctx).WithError(err).Warnf("target %s: %s cannot connect", target, address)
		}
		metricEntry := message.NewMessage("net_connectivity", message.Gauge, time.Now()).
			AddTag("target", target).
			AddField("value", value)
		metrics = append(metrics, metricEntry)
	}
	return metrics, nil
}
