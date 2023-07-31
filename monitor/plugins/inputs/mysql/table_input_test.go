package mysql

import (
	"context"
	"github.com/oceanbase/obagent/monitor/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCollect(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9878)/oceanbase?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          name: test
          minObVersion: ~
          maxObVersion: 4.0.0.0
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
	tableInput.Init(context.Background(), configMap)
	metrics, err := tableInput.CollectMsgs(context.Background())
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
	v, ok := tableInput.ConditionValueMap.Load("c1")
	require.True(t, !ok)
	cv, _ := utils.ConvertToFloat64(v)
	require.Equal(t, 0.0, cv)
}

func TestCollectSqlFailed(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9878)/oceanbase?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric1
          name: test
          minObVersion: ~
          maxObVersion: 4.0.0.0
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(context.Background(), configMap)
	metrics, err := tableInput.CollectMsgs(context.Background())
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectConvertFailed(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9878)/oceanbase?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric1
          name: test
          minObVersion: ~
          maxObVersion: 4.0.0.0
          tags:
            tag1: m1
            tag2: t2
          metrics:
            metric1: t1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(context.Background(), configMap)
	tableInput.Ob.MetaInfo.Version = "2.2.77"

	metrics, err := tableInput.CollectMsgs(context.Background())
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectConditionNotSatisfied(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9878)/oceanbase?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          name: test
          minObVersion: ~
          maxObVersion: 4.0.0.0
          condition: k0
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(context.Background(), configMap)
	metrics, err := tableInput.CollectMsgs(context.Background())
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectConditionNotFound(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9878)/oceanbase?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          name: test
          minObVersion: ~
          maxObVersion: 4.0.0.0
          condition: k2
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(context.Background(), configMap)
	metrics, err := tableInput.CollectMsgs(context.Background())
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestCollectConditionNotBool(t *testing.T) {
	tableInput := &TableInput{}

	config := `
      maintainCacheThreads: 4
      connection:
        url: user:pass@tcp(127.0.0.1:9878)/oceanbase?interpolateParams=true
        maxIdle: 2
        maxOpen: 32
      defaultConditionValues:
        k0: 0
        k1: 1
        k2: x
      collectConfig:
        - sql: select t1, t2, m1, m2 from test_metric
          name: test
          minObVersion: ~
          maxObVersion: 4.0.0.0
          condition: k2
          tags:
            tag1: t1
            tag2: t2
          metrics:
            metric1: m1
            metric2: m2
    `

	configMap, _ := utils.DecodeYaml(config)
	tableInput.Init(context.Background(), configMap)
	metrics, err := tableInput.CollectMsgs(context.Background())
	require.True(t, err == nil)
	require.Equal(t, 0, len(metrics))
}

func TestVersionSatisfied(t *testing.T) {
	unsatisfy := versionSatisfied("0.0.0", "4.0.0.0", "4.0.0.0")
	require.True(t, !unsatisfy)
	satisfy := versionSatisfied("0.0.0", "4.0.0.0", "3.2.3")
	require.True(t, satisfy)
	unsatisfyMinInfinity := versionSatisfied("", "4.0.0.0", "4.0.0.0")
	require.True(t, !unsatisfyMinInfinity)
	satisfyMinInfinity := versionSatisfied("", "4.0.0.0", "3.2.3")
	require.True(t, satisfyMinInfinity)
	unsatisfyMaxInfinity := versionSatisfied("4.0.0.0", "", "3.2.3")
	require.True(t, !unsatisfyMaxInfinity)
	satisfyMaxInfinity := versionSatisfied("4.0.0.0", "", "4.0.0.0")
	require.True(t, satisfyMaxInfinity)
	noVersion := versionSatisfied("", "", "3.2.3")
	require.True(t, noVersion)
}
