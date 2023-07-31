package web

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/api/common"
	http2 "github.com/oceanbase/obagent/lib/http"
)

type HttpServer struct {
	// server will be stopped, new request will be rejected
	Stopping int32
	// current session count, concurrent safely
	Counter *Counter
	// http routers
	Router      *gin.Engine
	LocalRouter *gin.Engine
	// address
	Address string
	// socket
	Socket string
	// http server, call its Run, Shutdown methods
	Server      *http.Server
	LocalServer *http.Server
	// stop the http.Server by calling cancel method
	Cancel context.CancelFunc
	// basic authorizer
	BasicAuthorizer common.Authorizer
}

func (server *HttpServer) AuthorizeMiddleware(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx).WithField("url", c.Request.URL)
	if server.BasicAuthorizer == nil {
		ctxlog.Warnf("basic auth is nil, please check the initial process.")
		c.Next()
		return
	}

	err := server.BasicAuthorizer.Authorize(c.Request)
	if err != nil {
		ctxlog.Errorf("basic auth Authorize failed, err:%+v", err)
		c.Abort()
		c.JSON(http.StatusUnauthorized, http2.BuildResponse(nil, err))
		return
	}
	c.Next()
}

// UseCounter use counter middleware
func (server *HttpServer) UseCounter() {
	server.Router.Use(
		server.counterPreHandlerFunc,
		server.counterPostHandlerFunc,
	)
}

// UseBasicAuth use basic auth middleware
func (server *HttpServer) UseBasicAuth() {
	server.Router.Use(
		server.AuthorizeMiddleware,
	)
}

// run start a httpServer
// when ctx is cancelled, call shutdown to stop the httpServer
func (server *HttpServer) Run(ctx context.Context) {

	server.Server.Handler = server.Router
	if server.Address != "" {
		tcpListener, err := http2.NewTcpListener(server.Address)
		if err != nil {
			log.WithError(err).
				Errorf("create tcp listener on address '%s' failed %v", server.Address, err)
			return
		}
		go func() {
			if err = server.Server.Serve(tcpListener); err != nil {
				log.WithError(err).
					Info("tcp server exited")
			}
		}()
	}
	if server.Socket != "" {
		socketListener, err := http2.NewSocketListener(server.Socket)
		if err != nil {
			log.WithError(err).
				Errorf("create socket listener on file '%s' failed %v", server.Socket, err)
			return
		}
		go func() {
			server.LocalServer = &http.Server{
				Handler:      server.LocalRouter,
				ReadTimeout:  5 * time.Minute,
				WriteTimeout: 5 * time.Minute,
			}
			if err = server.LocalServer.Serve(socketListener); err != nil {
				log.WithError(err).
					Info("socket server exited")
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			if err := server.Shutdown(ctx); err != nil {
				log.WithContext(ctx).
					WithError(err).
					Error("server shutdown failed!")
				// in a for loop, sleep 100ms
				time.Sleep(time.Millisecond * 100)
			} else {
				log.WithContext(ctx).Info("server shutdown successfully.")
				return
			}
		}
	}
}

// shutdown httpServer can shutdown if sessionCount is 0,
// otherwise, return an error
func (server *HttpServer) Shutdown(ctx context.Context) error {
	atomic.StoreInt32(&(server.Stopping), 1)
	sessionCount := atomic.LoadInt32(&server.Counter.sessionCount)
	if sessionCount > 0 {
		return errors.Errorf("server shutdown failed, cur-session count:%d, shutdown will be success when wait session-count is 0.", sessionCount)
	}
	return server.Server.Close()
}

// counterPreHandlerFunc middleware for httpServer session count, before process a request
func (server *HttpServer) counterPreHandlerFunc(c *gin.Context) {
	if atomic.LoadInt32(&(server.Stopping)) == 1 {
		c.Abort()
		c.JSON(http.StatusServiceUnavailable, http2.BuildResponse("server is shutdowning now.", nil))
		return
	}

	server.Counter.incr()

	c.Next()
}

// counterPostHandlerFunc middleware for httpServer session count, after process a request
func (server *HttpServer) counterPostHandlerFunc(c *gin.Context) {
	c.Next()

	server.Counter.decr()
}

// counter session counter
// when server receive a request, sessionCount +1,
// when the request returns a response, sessionCount -1.
type Counter struct {
	sessionCount int32
	sync.Mutex
}

// incr sessionCount +1 concurrent safely
func (c *Counter) incr() {
	c.Lock()
	c.sessionCount++
	defer c.Unlock()
}

// decr sessionCount -1 concurrent safely
func (c *Counter) decr() {
	c.Lock()
	c.sessionCount--
	defer c.Unlock()
}
