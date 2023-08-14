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
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/api/common"
	mgrroute "github.com/oceanbase/obagent/api/mgragent"
	"github.com/oceanbase/obagent/config"
	mgrconfig "github.com/oceanbase/obagent/config/mgragent"
	"github.com/oceanbase/obagent/executor/agent"
	http2 "github.com/oceanbase/obagent/lib/http"
	path2 "github.com/oceanbase/obagent/lib/path"
)

type Server struct {
	Config          mgrconfig.ServerConfig
	Router          *gin.Engine
	LocalRouter     *gin.Engine
	HttpServer      *http.Server
	LocalHttpServer *http.Server
	state           *http2.StateHolder
}

func NewServer(mode config.AgentMode, conf mgrconfig.ServerConfig) *Server {
	if mode == config.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	localRouter := gin.New()

	// TODO use gin.Logger() only for debugging, remove it before 3.2.0 publishes.
	if mode == config.DebugMode {
		router.Use(gin.Logger())
	}

	ret := &Server{
		Router:      router,
		LocalRouter: localRouter,
		Config:      conf,
		state:       http2.NewStateHolder(http2.Running),
	}
	router.Use(common.IgnoreFaviconHandler)
	router.Use(common.AuthorizeMiddleware)
	mgrroute.InitManagerAgentRoutes(ret.state, router)
	mgrroute.InitManagerAgentRoutes(ret.state, localRouter)
	common.InitPprofRouter(localRouter)
	return ret
}

func (s *Server) Run() {
	s.HttpServer = &http.Server{
		Handler:      s.Router,
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
	}
	s.LocalHttpServer = &http.Server{
		Handler:      s.LocalRouter,
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
	}
	tcpListener, err := http2.NewTcpListener(s.Config.Address)
	if err != nil {
		log.WithError(err).Fatalf("create tcp listener on %s", s.Config.Address)
	}
	socketPath := agent.SocketPath(s.Config.RunDir, path2.ProgramName(), os.Getpid())
	socketListener, err := http2.NewSocketListener(socketPath)
	if err != nil {
		log.WithError(err).Fatalf("create socket listener on %s", socketPath)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		err = s.HttpServer.Serve(tcpListener)
		if err != nil {
			log.WithError(err).Fatal("serve on tcp listener failed")
		}
		wg.Done()
	}()
	go func() {
		err = s.LocalHttpServer.Serve(socketListener)
		if err != nil {
			log.WithError(err).Fatal("serve on socket listener failed")
		}
		wg.Done()
	}()
	wg.Wait()
}

func (s *Server) Stop() {
	err := s.HttpServer.Shutdown(context.Background())
	log.WithError(err).Error("stop http server got error")
	s.state.Set(http2.Stopped)
	// TODO: wait command finished
	for mgrroute.TaskCount() > 0 {
		time.Sleep(time.Second)
	}
}

func (s *Server) State() http2.State {
	return s.state.Get()
}
