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

package route

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
)

// http handler: validate module config and save module config
func updateModuleConfigHandler(c *gin.Context) {
	kvs := config.KeyValues{}
	c.Bind(&kvs)
	ctx := NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx)

	configVersion, err := config.UpdateConfig(ctx, &kvs, true)
	if err != nil {
		ctxlog.Errorf("update config err:%+v", err)
	}

	sendResponse(c, configVersion, err)
}

// http handler: notify module config
func notifyModuleConfigHandler(c *gin.Context) {
	nconfig := new(config.NotifyModuleConfig)
	c.Bind(nconfig)

	ctx := NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx).WithFields(log.Fields{
		"process":            nconfig.Process,
		"module":             nconfig.Module,
		"updated key values": nconfig.UpdatedKeyValues,
	})

	ctxlog.Debugf("notify module config")

	err := config.NotifyModuleConfigForHttp(ctx, nconfig)
	if err != nil {
		ctxlog.Errorf("notify module config err:%+v", err)
	}

	sendResponse(c, "notify module config success", err)
}
