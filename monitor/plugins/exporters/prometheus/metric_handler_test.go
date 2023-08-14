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

package prometheus

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oceanbase/obagent/monitor/engine"
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/stretchr/testify/assert"
)

func TestPipelineGroupReader(t *testing.T) {
	msg1 := message.NewMessage("test1", message.Counter, time.Now()).AddTag("t", "1").AddField("count", 1.1)
	msg3 := message.NewMessage("test3", message.Counter, time.Now()).AddTag("t", "1").AddField("count", 1.1)
	cache1 := map[string]*message.Message{
		"test1": msg1,
		"test3": msg3,
	}

	msg2 := message.NewMessage("test2", message.Counter, time.Now()).AddTag("t", "2").AddField("count", 2.2)
	cache2 := map[string]*message.Message{
		"test1": msg1,
		"test2": msg2,
	}
	engine.GetRouteManager().AddPipelineGroup("/test/metric", cache1)
	engine.GetRouteManager().AddPipelineGroup("/test/metric", cache2)
	engine.GetRouteManager().RegisterHTTPRoute(context.Background(), "/test/metric", &httpRoute{routePath: "/test/metric"})

	ts := httptest.NewServer(&httpRoute{routePath: "/test/metric"})
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, `# HELP test1_count monitor collected message
# TYPE test1_count counter
test1_count{t="1"} 1.1
# HELP test2_count monitor collected message
# TYPE test2_count counter
test2_count{t="2"} 2.2
# HELP test3_count monitor collected message
# TYPE test3_count counter
test3_count{t="1"} 1.1
`, string(data))
}
