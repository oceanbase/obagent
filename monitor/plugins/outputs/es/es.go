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

package es

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"

	"github.com/lestrrat-go/strftime"
	es7 "github.com/opensearch-project/opensearch-go"
	esutil "github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/monitor/message"
)

type ESOutput struct {
	config  Config
	client  *es7.Client
	indexer esutil.BulkIndexer
	stat    Stat
}

type Config struct {
	ClientAddresses  string        `yaml:"clientAddresses"`
	Auth             Auth          `yaml:"auth"`
	IndexNamePattern string        `yaml:"indexNamePattern"`
	BatchSizeInBytes int           `yaml:"batchSizeInBytes"`
	MaxBatchWait     time.Duration `yaml:"maxBatchWait"`
	DocMap           DocMap        `yaml:"docMap"`
	RoutingField     string        `yaml:"routingField"`
	MaxRetries       int           `yaml:"maxRetries"`
}

type Auth struct {
	Username string `yaml:"username"` // Username for HTTP Basic Authentication.
	Password string `yaml:"password"` // Password for HTTP Basic Authentication.
}

type DocMap struct {
	Timestamp          string            `yaml:"timestamp"`
	TimestampPrecision time.Duration     `yaml:"timestampPrecision"`
	Name               string            `yaml:"name"`
	Tags               map[string]string `yaml:"tags"`
	Fields             map[string]string `yaml:"fields"`
}

type Stat struct {
	indexOk   int64
	indexFail int64
}

func (es *ESOutput) Description() string {
	return "es output"
}

func (es *ESOutput) SampleConfig() string {
	return ""
}

func NewESOutput(config map[string]interface{}) (*ESOutput, error) {
	es := &ESOutput{}
	var pluginConfig Config
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "ESOutput encode config")
	}
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return nil, errors.Wrap(err, "ESOutput decode config")
	}

	_, err = strftime.Format(pluginConfig.IndexNamePattern, time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "bad IndexNamePattern")
	}
	if pluginConfig.MaxRetries == 0 {
		pluginConfig.MaxRetries = 3
	}
	es.config = pluginConfig
	log.Infof("ESOutput init with config %v", pluginConfig)
	cfg := toESConfig(pluginConfig)
	client, err := es7.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	bulkIndexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:        client,
		FlushInterval: pluginConfig.MaxBatchWait,
		FlushBytes:    pluginConfig.BatchSizeInBytes,
		NumWorkers:    1,
		OnError: func(ctx context.Context, err error) {
			log.WithContext(ctx).WithError(err).Errorf("index error")
		},
	})
	if err != nil {
		return nil, err
	}
	es.client = client
	es.indexer = bulkIndexer
	return es, nil
}

func toESConfig(config Config) es7.Config {
	var esAddress []string
	for _, addr := range strings.Split(config.ClientAddresses, ",") {
		esAddress = append(esAddress, "http://"+addr)
	}
	return es7.Config{
		Addresses:     esAddress,
		Username:      config.Auth.Username,
		Password:      config.Auth.Password,
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			return time.Second
		},
		EnableRetryOnTimeout: true,
		MaxRetries:           config.MaxRetries,
		UseResponseCheckOnly: true,
		//DisableMetaHeader: true,
	}
}

func (es *ESOutput) Close() error {
	_ = es.indexer.Close(context.Background())
	return nil
}

func (es *ESOutput) Start(in <-chan []*message.Message) error {
	for messages := range in {
		err := es.Write(context.Background(), messages)
		if err != nil {
			log.WithError(err).Error("es write failed")
			return err
		}
	}
	return nil
}

func (es *ESOutput) Stop() {}

func (es *ESOutput) getRouting(msg *message.Message) *string {
	if es.config.RoutingField != "" {
		traceId, ok := msg.GetTag(es.config.RoutingField)
		if ok && traceId != "" {
			return &traceId
		}
	}
	return nil
}

func (es *ESOutput) Write(ctx context.Context, metrics []*message.Message) error {
	tmpDoc := make(map[string]interface{})
	for _, msg := range metrics {
		toDocMap(es.config.DocMap, msg, tmpDoc)
		docJson, err := json.Marshal(tmpDoc)
		if err != nil {
			log.WithError(err).Warnf("marshal docJson: %v failed", docJson)
			continue
		}
		indexName := es.indexName(msg.GetTime())
		bulkItem := es.toBulkItem(indexName, es.getRouting(msg), docJson)
		err = es.indexer.Add(ctx, bulkItem)
		if err != nil {
			return err // occurs when ctx.Done()
		}
	}
	return nil
}

func (es *ESOutput) indexName(t time.Time) string {
	ret, _ := strftime.Format(es.config.IndexNamePattern, t)
	return ret
}

func (es *ESOutput) toBulkItem(indexName string, routing *string, docJson []byte) esutil.BulkIndexerItem {
	return esutil.BulkIndexerItem{
		Action:  "index",
		Index:   indexName,
		Body:    bytes.NewReader(docJson),
		Routing: routing,
		OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem) {
			atomic.AddInt64(&es.stat.indexOk, 1)
		},
		OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem, err error) {
			atomic.AddInt64(&es.stat.indexFail, 1)
		},
	}
}

func toDocMap(config DocMap, msg *message.Message, docMap map[string]interface{}) {
	clearMap(docMap)
	if config.Name != "" {
		docMap[config.Name] = msg.GetName()
	}
	if config.Timestamp != "" {
		docMap[config.Timestamp] = msg.GetTime().UnixNano() / int64(config.TimestampPrecision)
	}
	for _, entry := range msg.Tags() {
		if to, ok := config.Tags[entry.Name]; ok {
			docMap[to] = entry.Value
		}
	}
	for _, entry := range msg.Fields() {
		if to, ok := config.Fields[entry.Name]; ok {
			docMap[to] = entry.Value
		}
	}
}

// go compiler will optimize it as a simple clear
func clearMap(m map[string]interface{}) {
	for k := range m {
		delete(m, k)
	}
}
