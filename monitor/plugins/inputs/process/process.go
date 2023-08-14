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

package process

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/plugins/common"
)

const sampleConfig = `
processNames: [observer, obproxy]
`

const description = `
collect process info
`

type ProcessConfig struct {
	ProcessNames    []string      `yaml:"processNames"`
	CollectInterval time.Duration `yaml:"collect_interval"`
}

type ProcessInput struct {
	Config *ProcessConfig
	Env    string

	ctx  context.Context
	done chan struct{}
}

func (p *ProcessInput) Init(ctx context.Context, config map[string]interface{}) error {
	var pluginConfig ProcessConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "process input encode config")
	}
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "process input decode config")
	}
	p.Config = &pluginConfig
	p.ctx = ctx
	p.done = make(chan struct{})

	env, err := common.CheckNodeEnv(ctx)
	if err != nil {
		return errors.Wrap(err, "check node env failed")
	}
	p.Env = env
	return nil
}

func (p *ProcessInput) Start(out chan<- []*message.Message) error {
	log.WithContext(p.ctx).Info("processInput started")
	go p.update(p.ctx, out)
	return nil
}

func (p *ProcessInput) update(ctx context.Context, out chan<- []*message.Message) {
	ticker := time.NewTicker(p.Config.CollectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			processMsgs, err := p.CollectMsgs(ctx)
			if err != nil {
				log.WithContext(ctx).Warnf("collect process messages failed, reason: %s", err)
			}
			out <- processMsgs
		case <-p.done:
			log.WithContext(ctx).Info("processInput exited")
			return
		}
	}
}

func (p *ProcessInput) Stop() {
	if p.done != nil {
		close(p.done)
	}
}

func (p *ProcessInput) SampleConfig() string {
	return sampleConfig
}

func (p *ProcessInput) Description() string {
	return description
}

func (p *ProcessInput) CollectMsgs(ctx context.Context) ([]*message.Message, error) {
	metrics := make([]*message.Message, 0, len(p.Config.ProcessNames))
	processes, err := allProcessNames()
	if err != nil {
		log.WithContext(ctx).Warnf("get all processes name failed, %s", err)
		return metrics, nil
	}

	for _, expectedName := range p.Config.ProcessNames {
		var value float64
		value = 0.0
		for _, processName := range processes {
			if processName == expectedName {
				value = 1.0
				break
			}
		}
		metricEntry := message.NewMessage("process_exists", message.Gauge, time.Now()).
			AddTag("name", expectedName).
			AddTag("env_type", p.Env).
			AddField("value", value)
		metrics = append(metrics, metricEntry)
	}
	return metrics, nil
}

var allProcessNames = func() ([]string, error) {
	allProcesses := common.GetProcesses()
	processNames := make([]string, 0, len(allProcesses.Processes))
	for _, proc := range allProcesses.Processes {
		processNames = append(processNames, proc.Name)
	}
	return processNames, nil
}
