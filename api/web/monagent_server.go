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

package web

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obagent/api/common"
	monroute "github.com/oceanbase/obagent/api/monagent"
	monconfig "github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/executor/agent"
	path2 "github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/monitor/engine"
)

var monitorAgentServer *MonitorAgentServer

func GetMonitorAgentServer() *MonitorAgentServer {
	return monitorAgentServer
}

type MonitorAgentServer struct {
	// original configs
	Config *monconfig.MonitorAgentConfig
	// sever of monitor metrics, selfstat, monitor manager API
	Server *HttpServer
	// two servers concurrent waitGroup
	wg *sync.WaitGroup
}

// NewMonitorAgentServer init monagent server: init configs and logger, register routers
func NewMonitorAgentServer(conf *monconfig.MonitorAgentConfig) *MonitorAgentServer {
	monagentServer := &MonitorAgentServer{
		Config: conf,
		Server: &HttpServer{
			Counter:         new(Counter),
			Router:          gin.New(),
			LocalRouter:     gin.New(),
			BasicAuthorizer: new(common.BasicAuth),
			Server:          &http.Server{},
			Address:         conf.Server.Address,
			Socket:          agent.SocketPath(conf.Server.RunDir, path2.ProgramName(), os.Getpid()),
		},
		wg: &sync.WaitGroup{},
	}
	// register middleware before register handlers
	monroute.UseMonitorMiddleware(monagentServer.Server.Router)
	monroute.UseLocalMonitorMiddleware(monagentServer.Server.LocalRouter)
	monitorAgentServer = monagentServer
	return monitorAgentServer
}

// Run start monagent servers: admin server, monitor server
func (server *MonitorAgentServer) Run() {
	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		ctx, cancel := context.WithCancel(context.Background())
		server.Server.Cancel = cancel
		server.Server.Run(ctx)
	}()
	server.wg.Wait()
}

// registerRouter register routers such as adminServer router and monitor metrics router
func (server *MonitorAgentServer) RegisterRouter() {
	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		server.Server.UseCounter()
		monroute.InitMonitorAgentRoutes(server.Server.Router, server.Server.LocalRouter)
		for router := range engine.PipelineRouteChan {
			monroute.RegisterPipelineRoute(router.Ctx, server.Server.Router, router.ExposeUrl, router.FuncHandler)
			monroute.RegisterPipelineRoute(router.Ctx, server.Server.LocalRouter, router.ExposeUrl, router.FuncHandler)
		}
	}()
}
