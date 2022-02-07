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

package engine

import (
	"context"
	"net/http"
	"sync"
    "time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/oceanbase/obagent/api/route"
	"github.com/oceanbase/obagent/api/web"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/utils"
)

var monitorAgentServer *MonitorAgentServer

func GetMonitorAgentServer() *MonitorAgentServer {
	return monitorAgentServer
}

type MonitorAgentServer struct {
	// original configs
	Config *config.MonitorAgentConfig
	// sever of monitor metrics, selfstat, monitor manager API
	Server *web.HttpServer
	// server of pprof
	AdminServer *web.HttpServer
	// server status map
	serverStatusMap sync.Map
}

// NewMonitorAgentServer init monagent server: init configs and logger, register routers
func NewMonitorAgentServer(conf *config.MonitorAgentConfig) *MonitorAgentServer {
	monagentServer := &MonitorAgentServer{
		Config: conf,
		Server: &web.HttpServer{
			Counter:         new(web.Counter),
			Router:          gin.Default(),
			BasicAuthorizer: new(web.BasicAuth),
			Server:          &http.Server{},
			Address:         conf.Server.Address,
		},
		AdminServer: &web.HttpServer{
			Counter:         new(web.Counter),
			Router:          gin.Default(),
			BasicAuthorizer: new(web.BasicAuth),
			Server:          &http.Server{},
			Address:         conf.Server.AdminAddress,
		},
	}
	monitorAgentServer = monagentServer
	return monitorAgentServer
}

// Run start mongagent servers: admin server, monitor server
func (server *MonitorAgentServer) Run() error {
	// check port available before start server
	go func() {
		server.serverStatusMap.Store("adminServer", false)
		ctx, cancel := context.WithCancel(context.Background())
		server.AdminServer.Cancel = cancel
		server.AdminServer.Run(ctx)
	}()
	go func() {
		server.serverStatusMap.Store("server", false)
		ctx, cancel := context.WithCancel(context.Background())
		server.Server.Cancel = cancel
		server.Server.Run(ctx)
	}()

	for {
		adminServerStatus, _ := server.serverStatusMap.LoadOrStore("adminServer", true)
		serverStatus, _ := server.serverStatusMap.LoadOrStore("server", true)
		adminServerOk, convertAdminServerOk := utils.ConvertToBool(adminServerStatus)
		serverOk, convertServerOk := utils.ConvertToBool(serverStatus)
		if !(convertAdminServerOk && convertServerOk) {
			return errors.New("start monagent server failed, adminSer")
		}
		if !(adminServerOk && serverOk) {
			return errors.New("start monagent server failed, adminSer")
		}
        time.Sleep(time.Second * 10)
	}
	return nil
}

// registerRouter register routers such as adminServer router and moniter metrics router
func (server *MonitorAgentServer) RegisterRouter() error {
	if err := server.registerAdminServerRouter(); err != nil {
		return errors.Wrap(err, "monitor agent server register admin server router")
	}
	if err := server.registerServerRouter(); err != nil {
		return errors.Wrap(err, "monitor agent server register metrics server router")
	}
	return nil
}

// registerServerRouter routers for moniter metrics.
func (server *MonitorAgentServer) registerServerRouter() error {
	server.Server.UseCounter()
	route.InitMonagentRoutes(server.Server.Router)
	return nil
}

// registerAdminServerRouter routers for selfstat, pprof and so on.
func (server *MonitorAgentServer) registerAdminServerRouter() error {
	server.AdminServer.UseCounter()
	route.InitPprofRouter(server.AdminServer.Router)
	return nil
}
