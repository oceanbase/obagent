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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/oceanbase/obagent/utils"
)

func TestCollect(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9876)/test?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          name: test
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
          conditionValues:
            c1: m1
    `
	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(configMap)
	metrics, err := tableInput.Collect()
	require.True(t, err == nil)
	require.Equal(t, 4, len(metrics))
	v, ok := tableInput.ConditionValueMap.Load("c1")
	require.True(t, ok)
	cv, _ := utils.ConvertToFloat64(v)
	require.Equal(t, 4.0, cv)
}

func TestCollectSqlFailed(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9876)/test?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric1
          name: test
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(configMap)
	metrics, err := tableInput.Collect()
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectConvertFailed(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9876)/test?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric1
          name: test
          tags:
            tag1: m1
            tag2: t2
          metrics:
            metric1: t1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(configMap)
	metrics, err := tableInput.Collect()
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectConditionNotSatisfied(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9876)/test?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          condition: k0
          name: test
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(configMap)
	metrics, err := tableInput.Collect()
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectConditionNotFound(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9876)/test?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          condition: k2
          name: test
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(configMap)
	metrics, err := tableInput.Collect()
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectConditionNotBool(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9876)/test?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
        k2: x
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          condition: k2
          name: test
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(configMap)
	metrics, err := tableInput.Collect()
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectEnableCache(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9876)/test?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          name: test
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
          enableCache: true
          cacheExpire: -60s
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(configMap)
	entry, found := tableInput.CacheMap.Load("test")
	require.True(t, !found)
	metrics, err := tableInput.Collect()
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
	for i := 0; i < 10 && !found; i++ {
		entry, found = tableInput.CacheMap.Load("test")
		time.Sleep(time.Second)
	}
	require.True(t, found)
	cachedEntry, ok := entry.(*CacheEntry)
	require.True(t, ok)
	require.Equal(t, 4, len(*cachedEntry.Metrics))
	tCachedFirst := cachedEntry.LoadTime
	tableInput.Collect()
	for i := 0; i < 10 && cachedEntry.LoadTime.Equal(tCachedFirst); i++ {
		entry, found = tableInput.CacheMap.Load("test")
		cachedEntry, _ = entry.(*CacheEntry)
		time.Sleep(time.Second)
	}
	require.True(t, cachedEntry.LoadTime.After(tCachedFirst))
}
