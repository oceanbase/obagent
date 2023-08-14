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

package mysql

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/errors"
	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/plugins/common"
	"github.com/oceanbase/obagent/monitor/utils"
	"github.com/oceanbase/obagent/stat"
)

const sampleConfig = `
maintainCacheThreads: 4
connection:
  url: user:password@tcp(127.0.0.1:2881)/oceanbase?interpolateParams=true
  maxIdle: 2
  maxOpen: 32
defaultConditionValues:
  key: value
collectConfig:
  - name: metricName
    tags:
      t1: c1
      t2: c2
    metrics:
      m1: c3
    conditionValues:
      key: c4
    subConfig:
      sql: select c1, c2, c3, c4 from t where c5=?
      params: [value1]
    enableCache: true
    cacheExpire: 10m
    cacheDataExpire: 20m
`

const description = `
collect data from database table
`

type TableCollectConfig struct {
	Name                    string            `yaml:"name"`
	Sql                     string            `yaml:"sql"`
	Params                  []string          `yaml:"params"`
	Condition               string            `yaml:"condition"`
	TagColumnMap            map[string]string `yaml:"tags"`
	MetricColumnMap         map[string]string `yaml:"metrics"`
	ConditionValueColumnMap map[string]string `yaml:"conditionValues"`
	EnableCache             bool              `yaml:"enableCache"`
	CacheExpire             time.Duration     `yaml:"cacheExpire"`
	SqlSlowThreshold        time.Duration     `yaml:"sqlSlowThreshold"`
	MinObVersion            string            `yaml:"minObVersion"`
	MaxObVersion            string            `yaml:"maxObVersion"`
}

type TableInputConfig struct {
	DbConnectionConfig       *common.DbConnectionConfig `yaml:"connection"`
	DefaultConditionValueMap map[string]interface{}     `yaml:"defaultConditionValues"`
	CollectConfigs           []*TableCollectConfig      `yaml:"collectConfig"`
	CollectInterval          time.Duration              `yaml:"collect_interval"`
	TimeAlign                bool                       `yaml:"timeAlign"`
}

type TableInput struct {
	Config            *TableInputConfig
	ConditionValueMap sync.Map
	Db                *sql.DB
	Ob                *common.Observer
	configLocker      sync.RWMutex

	ctx  context.Context
	done chan struct{}
}

func (t *TableInput) initDbConnection() error {
	db, err := sql.Open("mysql", t.Config.DbConnectionConfig.Url)
	if err != nil {
		return errors.Wrap(err, "db init")
	}
	db.SetMaxOpenConns(t.Config.DbConnectionConfig.MaxOpen)
	db.SetMaxIdleConns(t.Config.DbConnectionConfig.MaxIdle)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = db.PingContext(timeoutCtx)
	t.Db = db
	if err != nil {
		return errors.Wrap(err, "db ping")
	}
	return nil
}

func (t *TableInput) Close() error {
	if t.done != nil {
		close(t.done)
	}
	if t.Db != nil {
		return t.Db.Close()
	}
	return nil
}

func (t *TableInput) SampleConfig() string {
	return sampleConfig
}

func (t *TableInput) Description() string {
	return description
}

func (t *TableInput) Init(ctx context.Context, config map[string]interface{}) error {

	var pluginConfig TableInputConfig

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "mysqlTableInput encode config")
	}

	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "mysqlTableInput decode config")
	}

	t.Config = &pluginConfig
	t.ctx = context.Background()
	t.done = make(chan struct{})

	log.WithContext(ctx).Infof("init mysqlTableInput with config: %+v", t.Config)

	err = t.initDbConnection()
	if err != nil {
		return err
	}

	for k, v := range t.Config.DefaultConditionValueMap {
		t.ConditionValueMap.Store(k, v)
	}

	t.Ob, err = common.GetObserver(t.Config.DbConnectionConfig)
	if err != nil {
		return errors.Wrap(err, "get obVersion failed")
	}

	log.WithContext(ctx).Info("successfully init mysqlTableInput plugin")
	return nil
}

func (t *TableInput) Start(out chan<- []*message.Message) error {
	log.Infof("start tableInput plugin")
	go t.update(out)
	return nil
}

func (t *TableInput) update(out chan<- []*message.Message) {
	ticker := time.NewTicker(t.Config.CollectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			msgs, err := t.CollectMsgs(t.ctx)
			if err != nil {
				log.WithContext(t.ctx).Warnf("tableInput collect failed, reason: %s", err)
				continue
			}
			out <- msgs
		case <-t.done:
			log.Info("tableInput plugin exited")
			return
		}
	}
}

func (t *TableInput) Stop() {
	if t.done != nil {
		close(t.done)
	}
	if t.Db != nil {
		t.Db.Close()
	}
}

func (t *TableInput) doRecv(ctx context.Context, metrics *[]*message.Message, metricChan chan *message.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	for metricEntry := range metricChan {
		log.WithContext(ctx).Debugf("recv message, name=%s, time=%+v, fields-length=%d, fields=%+v, tags=%+v", metricEntry.GetName(), metricEntry.GetTime(), len(metricEntry.Fields()), metricEntry.Fields(), metricEntry.Tags())
		*metrics = append(*metrics, metricEntry)
	}
}

func (t *TableInput) deleteCollect(ctx context.Context, config *TableCollectConfig) {
	newConfigs := make([]*TableCollectConfig, 0, len(t.Config.CollectConfigs))
	for _, it := range t.Config.CollectConfigs {
		if it.Name != config.Name {
			newConfigs = append(newConfigs, it)
		}
	}

	log.WithContext(ctx).Warnf("delete config %s, sql %s", config.Name, config.Sql)
	t.configLocker.Lock()
	t.Config.CollectConfigs = newConfigs
	t.configLocker.Unlock()
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

func (t *TableInput) collectWithConfig(ctx context.Context, cancel context.CancelFunc, config *TableCollectConfig) ([]*message.Message, error) {
	var metrics []*message.Message
	currentTime := time.Now()

	args := make([]interface{}, 0, 2)
	for _, conditionValueName := range config.Params {
		value, found := t.ConditionValueMap.Load(conditionValueName)
		if !found {
			return nil, errors.Errorf("condition value %s not found", conditionValueName)
		}
		args = append(args, value)
	}

	var querySql string
	doCollect := false
	if versionSatisfied(config.MinObVersion, config.MaxObVersion, t.Ob.MetaInfo.Version) {
		querySql = config.Sql
		doCollect = true
	}

	// The current version is not collected
	if !doCollect {
		return nil, nil
	}
	tStart := time.Now()
	entry := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, tStart)).WithField("query sql", querySql)
	results, err := t.Db.QueryContext(ctx, querySql, args...)
	entry.Debug("execute end")
	duration := time.Now().Sub(tStart)
	stat.MonitorAgentTableInputHistogram.WithLabelValues(config.Name).Observe(duration.Seconds())
	if duration > config.SqlSlowThreshold {
		log.WithContext(ctx).Warnf("slow sql, name: %s, duration: %s (over %s), sql: %s", config.Name, duration, config.SqlSlowThreshold, querySql)
	}
	if err != nil {
		// 1146: Table xxx doesn't exist
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1146 {
			cancel()
			t.deleteCollect(ctx, config)
			log.WithContext(ctx).Warnf("collect %s, sql: %s, err: %s", config.Name, querySql, err)
			return nil, err
		}
		log.WithContext(ctx).Warnf("failed to do collect with sql %s, args:%+v, err: %s", querySql, args, err)
		return nil, err
	}
	defer results.Close()

	columns, err := results.Columns()
	if err != nil {
		return nil, errors.Errorf("get columns failed, err: %s", err)
	}
	columnNum := len(columns)
	values := make([]interface{}, columnNum)
	valuePtrs := make([]interface{}, columnNum)
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	var lastRow *map[string]interface{}

	for results.Next() {
		resultMap := make(map[string]interface{})
		lastRow = &resultMap
		err := results.Scan(valuePtrs...)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("sql results scan value failed")
			continue
		}
		for i, colName := range columns {
			colName := strings.ToLower(colName)
			resultMap[colName] = values[i]
		}
		fields := make([]message.FieldEntry, 0, len(config.MetricColumnMap))
		tags := make([]message.TagEntry, 0, len(config.TagColumnMap))
		for metricName, metricColumnName := range config.MetricColumnMap {
			metricValue, found := resultMap[metricColumnName]
			if found {
				v, convertOk := utils.ConvertToFloat64(metricValue)
				if !convertOk {
					log.WithContext(ctx).Warnf("can not convert value of %s %s to float64", metricColumnName, metricValue)
				} else {
					fields = append(fields, message.FieldEntry{Name: metricName, Value: v})
				}
			}
		}
		for tagName, tagColumnName := range config.TagColumnMap {
			tagValue, found := resultMap[tagColumnName]
			if found {
				v, convertOk := utils.ConvertToString(tagValue)
				if !convertOk {
					log.WithContext(ctx).Warnf("can not convert value of %s %s to string", tagColumnName, tagValue)
				} else {
					tags = append(tags, message.TagEntry{Name: tagName, Value: v})
				}
			}
		}
		metricEntry := message.NewMessageWithTagsFields(config.Name, message.Untyped, currentTime, tags, fields)
		metrics = append(metrics, metricEntry)
	}
	for conditionName, conditionColumnName := range config.ConditionValueColumnMap {
		if lastRow != nil {
			conditionValue, found := (*lastRow)[conditionColumnName]
			if found {
				t.ConditionValueMap.Store(conditionName, conditionValue)
			}
		}
	}
	return metrics, nil
}

func (t *TableInput) doCollect(ctx context.Context, config *TableCollectConfig, metricChan chan *message.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(config.Condition) > 0 {
		value, found := t.ConditionValueMap.Load(config.Condition)
		if !found {
			log.WithContext(ctx).Warn("Condition value not found:", config.Condition)
			return
		} else {
			conditionSatisfied, ok := utils.ConvertToBool(value)
			if !ok {
				log.WithContext(ctx).Warn("condition value convert failed")
				return
			}
			if !conditionSatisfied {
				log.WithContext(ctx).Warn("condition not satisfied, skip do collect for:", config.Name)
				return
			}
		}
	}

	entry := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, time.Now())).WithField("name", config.Name)
	metrics, err := t.collectWithConfig(ctx, func() {}, config)
	entry.Debug("collect table data end")
	if err != nil {
		entry.Warnf("collect table err: %+v", err)
	}
	for _, metric := range metrics {
		metricChan <- metric
	}
}

func (t *TableInput) CollectMsgs(ctx context.Context) ([]*message.Message, error) {
	wgCollect := &sync.WaitGroup{}
	wgRecv := &sync.WaitGroup{}
	metricChan := make(chan *message.Message, 100)
	metrics := make([]*message.Message, 0, 64)
	wgRecv.Add(1)
	go t.doRecv(ctx, &metrics, metricChan, wgRecv)

	t.configLocker.RLock()
	collectConfigs := t.Config.CollectConfigs
	t.configLocker.RUnlock()
	for _, collectConfig := range collectConfigs {
		wgCollect.Add(1)
		go t.doCollect(ctx, collectConfig, metricChan, wgCollect)
	}

	wgCollect.Wait()
	log.WithContext(ctx).Debug("mysqlTableInput do collect all done")
	close(metricChan)
	wgRecv.Wait()
	log.WithContext(ctx).Debug("mysqlTableInput do recv all done")
	return metrics, nil
}
