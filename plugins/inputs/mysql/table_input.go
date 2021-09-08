// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package mysql

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/metric"
	"github.com/oceanbase/obagent/utils"
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
  - sql: select c1, c2, c3, c4 from t where c5=?
    params: [value1]
    name: metricName
    tags:
      t1: c1
      t2: c2
    metrics:
      m1: c3
    conditionValues:
      key: c4
    enableCache: true
    cacheExpire: 10m
    cacheDataExpire: 20m
`

const description = `
collect data from database table
`

type DbConnectionConfig struct {
	Url     string `yaml:"url"`
	MaxOpen int    `yaml:"maxOpen"`
	MaxIdle int    `yaml:"maxIdle"`
}

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
	CacheDataExpire         time.Duration     `yaml:"cacheDataExpire"`
}

type TableInputConfig struct {
	SlowSqlThresholdMilliseconds int64                  `yaml:"slowSqlThresholdMilliseconds"`
	DbConnectionConfig           *DbConnectionConfig    `yaml:"connection"`
	DefaultConditionValueMap     map[string]interface{} `yaml:"defaultConditionValues"`
	CollectConfigs               []*TableCollectConfig  `yaml:"collectConfig"`
	MaintainCacheThreads         int                    `yaml:"maintainCacheThreads"`
}

type CacheEntry struct {
	LoadTime time.Time
	Metrics  *[]metric.Metric
}

type TableInput struct {
	Config                  *TableInputConfig
	ConditionValueMap       sync.Map
	CacheMap                sync.Map
	Db                      *sql.DB
	CacheTaskChan           chan *TableCollectConfig
	BackgroundTaskWaitGroup sync.WaitGroup
}

func (t *TableInput) initDbConnection() error {
	db, err := sql.Open("mysql", t.Config.DbConnectionConfig.Url)
	if err != nil {
		return errors.Wrap(err, "sql open")
	}
	db.SetMaxOpenConns(t.Config.DbConnectionConfig.MaxOpen)
	db.SetMaxIdleConns(t.Config.DbConnectionConfig.MaxIdle)
	err = db.Ping()
	if err != nil {
		return errors.Wrap(err, "db ping")
	}
	t.Db = db
	return nil
}

func (t *TableInput) Close() error {
	close(t.CacheTaskChan)
	t.BackgroundTaskWaitGroup.Wait()
	err := t.Db.Close()
	return err
}

func (t *TableInput) SampleConfig() string {
	return sampleConfig
}

func (t *TableInput) Description() string {
	return description
}

func (t *TableInput) Init(config map[string]interface{}) error {

	var pluginConfig TableInputConfig

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "table input encode config")
	}

	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "table input decode config")
	}

	t.Config = &pluginConfig

	log.Info("table input init with config", t.Config)

	nameMap := make(map[string]bool)
	for _, collectConfig := range pluginConfig.CollectConfigs {
		_, exists := nameMap[collectConfig.Name]
		if exists {
			return errors.Errorf("collect config %s already exists", collectConfig.Name)
		} else {
			nameMap[collectConfig.Name] = true
		}
	}

	err = t.initDbConnection()
	if err != nil {
		return errors.Wrap(err, "db init connection")
	}

	for k, v := range t.Config.DefaultConditionValueMap {
		t.ConditionValueMap.Store(k, v)
	}

	t.CacheTaskChan = make(chan *TableCollectConfig, 100)
	for i := 0; i < t.Config.MaintainCacheThreads; i++ {
		t.BackgroundTaskWaitGroup.Add(1)
		go t.doUpdateCache()
	}

	log.Info("successfully init mysql table input plugin")
	return nil
}

func (t *TableInput) doRecv(metrics *[]metric.Metric, metricChan chan metric.Metric, wg *sync.WaitGroup) {
	defer wg.Done()
	for metricEntry := range metricChan {
		log.Debug("recv get metric:", metricEntry.GetName(), metricEntry.GetTime(), metricEntry.Fields(), len(metricEntry.Fields()), metricEntry.Tags())
		*metrics = append(*metrics, metricEntry)
	}
}

func (t *TableInput) doUpdateCache() {
	defer t.BackgroundTaskWaitGroup.Done()
	for config := range t.CacheTaskChan {
		log.Info("got update cache task for metric:", config.Name)
		metrics := t.collectData(config)
		cacheEntry := &CacheEntry{
			LoadTime: time.Now(),
			Metrics:  &metrics,
		}
		t.CacheMap.Store(config.Name, cacheEntry)
	}
}

func (t *TableInput) collectData(config *TableCollectConfig) []metric.Metric {
	var metrics []metric.Metric
	currentTime := time.Now()
	args := make([]interface{}, len(config.Params))
	for idx, conditionValueName := range config.Params {
		value, found := t.ConditionValueMap.Load(conditionValueName)
		if !found {
			log.Warn("condition value not found should return:", conditionValueName)
			return metrics
		}
		args[idx] = value
	}
	tStart := time.Now()
	results, err := t.Db.Query(config.Sql, args...)
	tEnd := time.Now()
	if (tEnd.UnixNano()-tStart.UnixNano())/1000000 > t.Config.SlowSqlThresholdMilliseconds {
		log.Warnf("slow collect sql: %s", config.Sql)
	}
	if err != nil {
		log.Warn("failed to do collect with sql:", config.Sql, args, err)
		return metrics
	}
	defer results.Close()

	columns, err := results.Columns()
	if err != nil {
		log.Warn("get columns failed")
		return metrics
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
			log.WithError(err).Error("sql results scan value failed")
			continue
		}
		for i, colName := range columns {
			resultMap[colName] = values[i]
		}
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		for metricName, metricColumnName := range config.MetricColumnMap {
			metricValue, found := resultMap[metricColumnName]
			if found {
				v, convertOk := utils.ConvertToFloat64(metricValue)
				if !convertOk {
					log.Errorf("can not convert value of %s %s to float64", metricColumnName, metricValue)
				} else {
					fields[metricName] = v

				}
			}
		}
		for tagName, tagColumnName := range config.TagColumnMap {
			tagValue, found := resultMap[tagColumnName]
			if found {
				v, convertOk := utils.ConvertToString(tagValue)
				if !convertOk {
					log.Errorf("can not convert value of %s %s to string", tagColumnName, tagValue)
				} else {
					tags[tagName] = v
				}
			}
		}
		metricEntry := metric.NewMetric(config.Name, fields, tags, currentTime, metric.Untyped)
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
	return metrics
}

func (t *TableInput) doCollect(config *TableCollectConfig, metricChan chan metric.Metric, wg *sync.WaitGroup) {
	defer wg.Done()
	currentTime := time.Now()
	if len(config.Condition) > 0 {
		value, found := t.ConditionValueMap.Load(config.Condition)
		if !found {
			log.Warn("Condition value not found:", config.Condition)
			return
		} else {
			conditionSatisfied, ok := utils.ConvertToBool(value)
			if !ok {
				log.Warn("condition value convert failed")
				return
			}
			if !conditionSatisfied {
				log.Warn("condition not satisfied, skip do collect for:", config.Name)
				return
			}
		}
	}

	if config.EnableCache {
		entry, found := t.CacheMap.Load(config.Name)
		if found {
			cacheEntry := entry.(*CacheEntry)
			log.Debug("found value in cache for metric:", config.Name)
			if cacheEntry.LoadTime.Add(config.CacheDataExpire).After(currentTime) {
				for _, metric := range *cacheEntry.Metrics {
					metricChan <- metric.Clone()
				}
			}
			if cacheEntry.LoadTime.Add(config.CacheExpire).Before(currentTime) {
				log.Debug("cache expire, trigger update task:", config.Name)
				t.CacheTaskChan <- config
			}
		} else {
			log.Info("no value found in cache, trigger update cache:", config.Name)
			t.CacheTaskChan <- config
		}
	} else {
		metrics := t.collectData(config)
		for _, metric := range metrics {
			metricChan <- metric
		}
	}
}

func (t *TableInput) Collect() ([]metric.Metric, error) {
	wgCollect := &sync.WaitGroup{}
	wgRecv := &sync.WaitGroup{}
	metricChan := make(chan metric.Metric, 100)
	metrics := make([]metric.Metric, 0, 64)
	wgRecv.Add(1)
	go t.doRecv(&metrics, metricChan, wgRecv)
	for _, collectConfig := range t.Config.CollectConfigs {
		wgCollect.Add(1)
		go t.doCollect(collectConfig, metricChan, wgCollect)
	}
	wgCollect.Wait()
	log.Debug("table input do collect all done")
	close(metricChan)
	wgRecv.Wait()
	log.Debug("table input do recv all done")
	return metrics, nil
}
