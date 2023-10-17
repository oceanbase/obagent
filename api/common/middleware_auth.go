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

package common

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/errors"
	http2 "github.com/oceanbase/obagent/lib/http"
)

const (
	Basic           string = "Basic"
	HmacSHA256      string = "OCP-HMACSHA256"
	TRACE_ID_HEADER string = "X-OCP-Trace-ID"
	TimeFormat      string = "2006/01/02 15:04:05"
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
		log.Errorf("invalid header authorization: %s, should contain 2 content. url: %s", authHeader, req.URL)
		return errors.Errorf("invalid header authorization")
	}

	authType := authHeaders[0]
	if authType != Basic && authType != HmacSHA256 {
		log.Errorf("invalid header authorization: %s, first content should be Basic or OCP-HMACSHA256, got %s.", authHeader, authHeaders[0])
		return errors.Errorf("invalid header authorization")
	}

	switch authType {
	case Basic:
		content, err := base64.StdEncoding.DecodeString(authHeaders[1])
		if err != nil {
			log.Errorf("invalid header authorization: %s, decode base64 err: %s", authHeaders[1], err)
			return errors.Errorf("invalid header authorization")
		}

		// base64-encoding-content decode result: username:password
		contentStrs := strings.SplitN(string(content), ":", 2)
		if len(contentStrs) != 2 {
			log.Errorf("invalid header authorization: %s, decode base64 result err: %s does not contain :", authHeaders[1], content)
			return errors.Errorf("invalid header authorization")
		}

		if contentStrs[0] == auth.config.Username && contentStrs[1] == auth.config.Password {
			return nil
		} else {
			return errors.Errorf("wrong password")
		}
	case HmacSHA256:
		if !checkReqTime(req) {
			log.Errorf("check request Date failed")
			return errors.Occur(errors.ErrBadRequest)
		}
		// digest-content: username:signature
		contentStrs := strings.SplitN(authHeaders[1], ":", 2)
		if len(contentStrs) != 2 {
			log.Errorf("invalid header authorization: %s", authHeaders[1])
			return errors.Errorf("invalid header authorization")
		}
		sign := computeSign(req, auth.config.Password)
		if contentStrs[0] == auth.config.Username && contentStrs[1] == sign {
			return nil
		} else {
			log.Errorf("digest http auth failed, ocp-server-user: %s, agent-user: %s, sign is invalid.", contentStrs[0], auth.config.Username)
			return errors.Errorf("wrong digest")
		}
	}

	return errors.Errorf("no authentication")
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

func checkReqTime(req *http.Request) bool {
	date := req.Header.Get("Date")
	if date == "" {
		return true
	}
	timeParse, err := time.Parse(TimeFormat, date)
	if err != nil {
		log.WithError(err).Errorf("invalid request date, date: %s", date)
		return false
	}
	if time.Now().Sub(timeParse) > 60*time.Second {
		log.Warnf("request date is too early, just skip")
		return false
	}
	return true
}

func computeSign(req *http.Request, password string) string {
	method := req.Method
	url := req.URL.String()
	contentType := req.Header.Get("Content-Type")
	date := req.Header.Get("Date")
	traceId := req.Header.Get(TRACE_ID_HEADER)
	string2Sign := method + "\n" + url + "\n" + contentType + "\n" + date + "\n" + traceId
	mac := hmac.New(sha256.New, []byte(password))
	mac.Write([]byte(string2Sign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return signature
}
