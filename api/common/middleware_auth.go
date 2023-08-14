/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package common

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	http2 "github.com/oceanbase/obagent/lib/http"
)

type Authorizer interface {
	Authorize(req *http.Request) error
	SetConf(conf config.BasicAuthConfig)
}

func InitBasicAuthConf(ctx context.Context) {
	httpAuthorizer = &BasicAuth{}

	module := config.ManagerAgentBasicAuthConfigModule
	err := config.InitModuleConfig(ctx, module)
	if err != nil {
		log.WithContext(ctx).Fatalf("init module %s config err:%+v", module, err)
	}
	log.WithContext(ctx).Infof("init module %s config end", module)
}

var httpAuthorizer Authorizer

func NotifyConf(conf config.BasicAuthConfig) {
	httpAuthorizer.SetConf(conf)
}

type BasicAuth struct {
	config config.BasicAuthConfig
}

func (auth *BasicAuth) SetConf(conf config.BasicAuthConfig) {
	auth.config = conf
}

func (auth *BasicAuth) Authorize(req *http.Request) error {
	if !auth.config.MetricAuthEnabled && strings.HasPrefix(req.RequestURI, "/metrics/") {
		return nil
	}
	// header: Authorization Basic base64-encoding-content
	authHeader := req.Header.Get("Authorization")
	authHeaders := strings.SplitN(authHeader, " ", 2)
	if len(authHeaders) != 2 {
		return errors.Errorf("invalid header authorization: %s, should contain 2 content.", authHeader)
	}

	if authHeaders[0] != "Basic" {
		return errors.Errorf("invalid header authorization: %s, first content should be Basic, got %s.", authHeader, authHeaders[0])
	}
	var content []byte
	var err error
	if content, err = base64.StdEncoding.DecodeString(authHeaders[1]); err != nil {
		return errors.Errorf("invalid header authorization: %s, decode base64 err: %s", authHeaders[1], err)
	}

	// base64-encoding-content decode result: username:password
	contentStrs := strings.SplitN(string(content), ":", 2)
	if len(contentStrs) != 2 {
		return errors.Errorf("invalid header authorization: %s, decode base64 result err: %s does not contain :", authHeaders[1], content)
	}

	if contentStrs[0] == auth.config.Username && contentStrs[1] == auth.config.Password {
		return nil
	}

	return errors.Errorf("auth failed for user: %s", contentStrs[0])
}

func AuthorizeMiddleware(c *gin.Context) {
	ctx := NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx)
	if httpAuthorizer == nil {
		ctxlog.Warnf("basic auth is nil, please check the initial process.")
		c.Next()
		return
	}

	err := httpAuthorizer.Authorize(c.Request)
	if err != nil {
		ctxlog.Errorf("basic auth Authorize failed, err:%+v", err)
		c.Abort()
		c.JSON(http.StatusUnauthorized, http2.BuildResponse(nil, err))
		return
	}
	c.Next()
}
