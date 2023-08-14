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

package jointable

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/didi/gendry/scanner"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/oceanbase/obagent/errors"
	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/plugins/common"
)

const description = ``

const sampleConfig = `
connection:
  url: user:password@tcp(127.0.0.1:2881)/oceanbase?interpolateParams=true
  maxIdle: 2
  maxOpen: 32
joinTableConfigs:
  queryDataConfigs:
  - querySql: select 'tv1' as t1, 'tv2' as t2, 'tv3' as t3
    queryArgs: []
    minOBVersion: 4.0.0.0
  cacheExpire: 3m
  conditions:
  - metrics: [test]
    tags:
      t1: tv1
    tagNames: ["t1", "t2"]
    removeNotMatchedTagValueMessage: true
`

type JoinTablePluginConfig struct {
	DbConnectionConfig common.DbConnectionConfig `yaml:"connection"`
	JoinTableConfigs   []JoinTableConfig         `yaml:"joinTableConfigs"`
}

type JoinTableConfig struct {
	QueryDataConfigs []QueryDataConfig `yaml:"queryDataConfigs"`
	CacheExpire      time.Duration     `yaml:"cacheExpire"`

	Conditions []Condition `yaml:"conditions"`
}

type QueryDataConfig struct {
	QuerySQL     string        `yaml:"querySql"`
	QueryArgs    []interface{} `yaml:"queryArgs"`
	MinOBVersion string        `yaml:"minOBVersion"`
	MaxOBVersion string        `yaml:"maxOBVersion"`
}

type JoinTable struct {
	Config *JoinTablePluginConfig
	Db     *sql.DB
	Ob     *common.Observer
	Cache  *common.Cache

	ctx  context.Context
	done chan struct{}
}

func (c *JoinTablePluginConfig) init() {
	for _, conf := range c.JoinTableConfigs {
		conf.init()
	}
}

func (c *JoinTableConfig) init() {
	for _, cond := range c.Conditions {
		cond.init()
	}
}

func (j *JoinTable) initDbConnection() error {
	db, err := sql.Open("mysql", j.Config.DbConnectionConfig.Url)
	if err != nil {
		return errors.Wrap(err, "db init")
	}
	db.SetMaxOpenConns(j.Config.DbConnectionConfig.MaxOpen)
	db.SetMaxIdleConns(j.Config.DbConnectionConfig.MaxIdle)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = db.PingContext(timeoutCtx)
	j.Db = db
	if err != nil {
		return errors.Wrap(err, "db ping")
	}
	return nil
}

func (j *JoinTable) SampleConfig() string {
	return sampleConfig
}

func (j *JoinTable) Description() string {
	return description
}

func (j *JoinTable) Init(ctx context.Context, config map[string]interface{}) error {
	var pluginConfig JoinTablePluginConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "JoinTablePluginConfig encode config")
	}
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "JoinTablePluginConfig decode config")
	}

	j.Config = &pluginConfig
	j.Config.init()

	j.ctx = context.Background()
	j.done = make(chan struct{})

	log.WithContext(ctx).Infof("init JoinTablePluginConfig with config: %+v", j.Config)

	err = j.initDbConnection()
	if err != nil {
		return err
	}

	j.Ob, err = common.GetObserver(&j.Config.DbConnectionConfig)
	if err != nil {
		return errors.Wrap(err, "get obVersion failed")
	}

	ctxCancel, cancel := context.WithCancel(ctx)
	j.Cache = &common.Cache{
		Cancel: cancel,
	}
	for i, joinTableConfig := range pluginConfig.JoinTableConfigs {
		tmpConfig := joinTableConfig
		collectCtx, collectCancel := context.WithCancel(ctxCancel)

		loadFunc := func() (interface{}, error) {
			return j.collectDBData(ctx, collectCancel, tmpConfig)
		}
		go j.Cache.Update(collectCtx, fmt.Sprintf("jointable-%d", i), tmpConfig.CacheExpire, loadFunc)
	}
	log.WithContext(ctx).Info("successfully init mysqlTableInput plugin")
	return nil
}

func (j *JoinTable) Start(in <-chan []*message.Message, out chan<- []*message.Message) (err error) {
	for messages := range in {
		outMessages, err := j.Process(context.Background(), messages...)
		if err != nil {
			log.Warnf("jointable for message failed: %v", err)
		}
		out <- outMessages
	}
	return nil
}

func (j *JoinTable) Stop() {
	if j.done != nil {
		close(j.done)
	}

	if j.Cache != nil {
		j.Cache.Close()
	}
	if j.Db != nil {
		j.Db.Close()
	}
}

func (j *JoinTable) collectDBData(ctx context.Context, cancel context.CancelFunc, joinTableConfig JoinTableConfig) ([]map[string]string, error) {
	var querySQL string
	queryArgs := make([]interface{}, 0, 2)
	for _, dbDataConfig := range joinTableConfig.QueryDataConfigs {
		if versionSatisfied(dbDataConfig.MinOBVersion, dbDataConfig.MaxOBVersion, j.Ob.MetaInfo.Version) {
			querySQL = dbDataConfig.QuerySQL
			queryArgs = dbDataConfig.QueryArgs
			break
		}
	}
	if querySQL == "" {
		return nil, nil
	}

	tStart := time.Now()
	entry := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, tStart)).WithField("query sql", querySQL)
	results, err := j.Db.QueryContext(ctx, querySQL, queryArgs...)
	entry.Debug("execute end")
	duration := time.Now().Sub(tStart)
	// stat.MonitorAgentTableInputHistogram.WithLabelValues(config.Name).Observe(duration.Seconds())
	if duration > time.Millisecond*100 {
		log.WithContext(ctx).Warnf("slow sql, duration: %s (over 100ms), sql: %s", duration, querySQL)
	}
	if err != nil {
		log.WithContext(ctx).Warnf("failed to do collect with sql %s, args:%+v, err: %s", querySQL, queryArgs, err)
		return nil, err
	}
	defer results.Close()

	datas, err := scanner.ScanMapDecode(results)
	if err != nil {
		return nil, err
	}
	dbData := make([]map[string]string, 0, len(datas))
	for _, data := range datas {
		dbData = append(dbData, parseDataToTags(data))
	}

	return dbData, nil
}

func parseDataToTags(data map[string]interface{}) map[string]string {
	ret := make(map[string]string, len(data))
	for key, value := range data {
		ret[key] = fmt.Sprint(value)
	}
	return ret
}

func (j *JoinTable) Process(ctx context.Context, messages ...*message.Message) ([]*message.Message, error) {
	output := make([]*message.Message, 0, len(messages))
	for _, msg := range messages {
		if joinMsg := j.joinTable(ctx, msg, j.Config.JoinTableConfigs); joinMsg != nil {
			output = append(output, msg)
		}
	}
	return output, nil
}

func (j *JoinTable) joinTable(ctx context.Context, msg *message.Message, confs []JoinTableConfig) *message.Message {
	for i, joinTableConfig := range confs {
		cacheName := fmt.Sprintf("jointable-%d", i)
		entry, ex := j.Cache.CacheStore.Load(cacheName)
		if !ex {
			log.Warnf("cache %s not exist!", cacheName)
			continue
		}
		dbData, ok := entry.([]map[string]string)
		if !ok {
			log.WithContext(ctx).Warnf("type of value in cache for dbData %s is not map[string]string", cacheName)
			continue
		}

		for _, condition := range joinTableConfig.Conditions {
			removed := operateMessage(condition, msg, dbData)
			if removed {
				return nil
			}
		}
	}
	return msg
}

func versionSatisfied(minVersion string, maxVersion string, obVersion string) bool {
	if len(minVersion) != 0 {
		compareToMin, _ := common.CompareVersion(obVersion, minVersion)
		if compareToMin < 0 {
			return false
		}
	}
	if len(maxVersion) != 0 {
		compareToMax, _ := common.CompareVersion(obVersion, maxVersion)
		if compareToMax >= 0 {
			return false
		}
	}

	return true
}
