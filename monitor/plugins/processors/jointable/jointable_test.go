package jointable

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/auth"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/plugins/common"
)

func TestJoinTable(t *testing.T) {
	configStr := `
connection:
  url: user:pass@tcp(127.0.0.1:9878)/oceanbase?interpolateParams=true
  maxIdle: 2
  maxOpen: 32
joinTableConfigs:
  - queryDataConfigs:
    - querySql: select 'tv1' as t1, 'tv2' as t2, 3 as t3, true as t4
      queryArgs: []
      maxOBVersion: 4.0.0.0
    cacheExpire: 1m
    conditions:
    - metrics: [test]
      tags:
        t1: tv1
      tagNames: ["t1", "t2"]
      removeNotMatchedTagValueMessage: true
`
	processor := &JoinTable{}
	defer processor.Stop()

	var configMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &configMap)
	var pluginConfig JoinTablePluginConfig
	configBytes, err := yaml.Marshal(configMap)
	assert.Nil(t, err)
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	assert.Nil(t, err)
	processor.Config = &pluginConfig
	processor.Config.init()
	err = processor.initDbConnection()
	assert.Nil(t, err)

	processor.Ob = &common.Observer{}
	processor.Ob.MetaInfo = &common.ObserverMetaInfo{Version: "2.2.77"}

	ctxCancel, cancel := context.WithCancel(context.Background())
	processor.Cache = &common.Cache{
		Cancel: cancel,
	}
	for i, joinTableConfig := range pluginConfig.JoinTableConfigs {
		tmpConfig := joinTableConfig
		collectCtx, collectCancel := context.WithCancel(ctxCancel)

		loadFunc := func() (interface{}, error) {
			return processor.collectDBData(context.Background(), collectCancel, tmpConfig)
		}
		go processor.Cache.Update(collectCtx, fmt.Sprintf("jointable-%d", i), tmpConfig.CacheExpire, loadFunc)
	}
	time.Sleep(time.Millisecond * 100)

	in := make(chan []*message.Message, 2)
	out := make(chan []*message.Message, 2)
	go processor.Start(in, out)

	msg := message.NewMessage("test", message.Counter, time.Now())
	msg.AddTag("t1", "tv1")
	msg.AddTag("t2", "tv2")
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)
	in <- []*message.Message{msg}

	outmsg := <-out
	assert.Equal(t, 4, len(outmsg[0].Tags()))
	t3Value, ex := outmsg[0].GetTag("t3")
	assert.Equal(t, true, ex)
	assert.Equal(t, "3", t3Value)
	t4Value, ex := outmsg[0].GetTag("t4")
	assert.Equal(t, true, ex)
	assert.Equal(t, "1", t4Value)
}

func TestJoinTable_WithMultiData(t *testing.T) {
	configStr := `
connection:
  url: user:pass@tcp(127.0.0.1:9878)/oceanbase?interpolateParams=true
  maxIdle: 2
  maxOpen: 32
joinTableConfigs:
  - queryDataConfigs:
    - querySql: select 'tv1' as t1, 'tv2' as t2, 3 as t3, true as t4
      queryArgs: []
      minOBVersion: 100000.0.0.0
    cacheExpire: 1m
    conditions:
    - metrics: [test]
      tags:
        t1: tv1
      tagNames: [t1]
      removeNotMatchedTagValueMessage: true
`
	processor := &JoinTable{}
	defer processor.Stop()

	var configMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &configMap)
	var pluginConfig JoinTablePluginConfig
	configBytes, err := yaml.Marshal(configMap)
	assert.Nil(t, err)
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	assert.Nil(t, err)
	processor.Config = &pluginConfig
	processor.Config.init()
	err = processor.initDbConnection()
	assert.Nil(t, err)

	processor.Ob = &common.Observer{}
	processor.Ob.MetaInfo = &common.ObserverMetaInfo{Version: "2.2.77"}

	_, cancel := context.WithCancel(context.Background())
	processor.Cache = &common.Cache{
		Cancel: cancel,
	}
	processor.Cache.CacheStore.Store("jointable-0", []map[string]string{
		{
			"t1": "tv1-not-equal",
			"t2": "ne-tv2",
			"t3": "ne-3",
			"t4": "ne-4",
		},
		{
			"t1": "tv1",
			"t2": "tv2",
			"t3": "3",
			"t4": "4",
		},
		{
			"t1": "tv1-not-equal2",
			"t2": "ne2-tv2",
			"t3": "ne2-3",
			"t4": "ne2-4",
		},
	})

	in := make(chan []*message.Message, 2)
	out := make(chan []*message.Message, 2)
	go processor.Start(in, out)

	msg := message.NewMessage("test", message.Counter, time.Now())
	msg.AddTag("t1", "tv1")
	msg.AddTag("t2", "tv2")
	msg.AddField("f1", 1.0)
	msg.AddField("f2", 2.0)
	in <- []*message.Message{msg}

	outmsg := <-out
	assert.Equal(t, 4, len(outmsg[0].Tags()))
	t3Value, ex := outmsg[0].GetTag("t3")
	assert.Equal(t, true, ex)
	assert.Equal(t, "3", t3Value)
	t4Value, ex := outmsg[0].GetTag("t4")
	assert.Equal(t, true, ex)
	assert.Equal(t, "4", t4Value)
}

func TestMain(m *testing.M) {
	s := setup()
	code := m.Run()
	s.Close()
	os.Exit(code)
}

func setup() *server.Server {
	driver := sqle.NewDefault()
	driver.AddDatabase(memory.NewDatabase("oceanbase"))

	config := server.Config{
		Protocol: "tcp",
		Address:  "0.0.0.0:9878",
		Version:  "2.2.77",
		Auth:     auth.NewNativeSingle("user", "pass", auth.AllPermissions),
	}

	s, err := server.NewDefaultServer(config, driver)
	if err != nil {
		panic(err)
	}

	go s.Start()
	return s
}
