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

package obcommon

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/lib/mask"
	"github.com/oceanbase/obagent/lib/trace"
	"github.com/oceanbase/obagent/monitor/message"
)

const connectivitySampleConfig = `
targets:
  t1: 'user:password@tcp(127.0.0.1:2881)/oceanbase?interpolateParams=true'
`

const connectivityDescription = `
check oceanbase connectivity
`

type ConnectivityConfig struct {
	Targets         map[string]string `yaml:"targets"`
	CollectInterval time.Duration     `yaml:"collect_interval"`
}

type ConnectivityInput struct {
	Config *ConnectivityConfig
	Dbs    map[string]*sql.DB
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

	c.Dbs = make(map[string]*sql.DB)
	err = c.initDbConnections()
	if err != nil {
		return errors.Wrap(err, "init DB connections failed")
	}
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
				log.WithContext(ctx).Warnf("collect connectivity messages failed, reason: %s", err)
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
	for _, db := range c.Dbs {
		if db != nil {
			db.Close()
		}
	}
}

func (c *ConnectivityInput) SampleConfig() string {
	return connectivitySampleConfig
}

func (c *ConnectivityInput) Description() string {
	return connectivityDescription
}

func (c *ConnectivityInput) CollectMsgs() ([]*message.Message, error) {
	metrics := make([]*message.Message, 0, len(c.Config.Targets))
	ctx := trace.ContextWithRandomTraceId()
	for target := range c.Config.Targets {
		value := 0.0
		var selectRes int
		if db, ok := c.Dbs[target]; ok {
			row := db.QueryRow("select 1")
			err := row.Scan(&selectRes)
			if err == nil && selectRes == 1 {
				value = 1.0
			} else {
				log.WithContext(ctx).Warnf("target %s execute sql failed, err: %s", mask.Mask(target), err)
			}
		}

		metricEntry := message.NewMessage("oceanbase_connectivity", message.Gauge, time.Now()).
			AddTag("target", target).
			AddField("value", value)
		metrics = append(metrics, metricEntry)
	}
	return metrics, nil
}

func (c *ConnectivityInput) initDbConnections() error {
	for target, url := range c.Config.Targets {
		db, err := sql.Open("mysql", url)
		if err != nil {
			log.WithError(err).Errorf("target %s cannot open", mask.Mask(target))
			return err
		}
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err = db.PingContext(timeoutCtx)
		if err != nil {
			db.Close()
			log.WithError(err).Errorf("target %s ping failed", mask.Mask(target))
			return err
		}
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		c.Dbs[target] = db
	}
	return nil
}
