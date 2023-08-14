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
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/mgragent"
	"github.com/oceanbase/obagent/executor/agent"
)

// http handler: validate module config and save module config
func UpdateConfigPropertiesHandler(c *gin.Context) {
	kvs := config.KeyValues{}
	c.Bind(&kvs)
	ctx := NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx)

	configVersion, err := config.UpdateConfig(ctx, &kvs)
	if err != nil {
		ctxlog.Errorf("update config err:%s", err)
	}

	SendResponse(c, &agent.AgentctlResponse{
		Successful: true,
		Message:    configVersion,
		Error:      "",
	}, err)
}

// http handler: notify module config
func NotifyConfigPropertiesHandler(c *gin.Context) {
	nconfig := new(config.NotifyModuleConfig)
	c.Bind(nconfig)

	ctx := NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx).WithFields(log.Fields{
		"process":            nconfig.Process,
		"module":             nconfig.Module,
		"updated key values": nconfig.UpdatedKeyValues,
	})

	ctxlog.Infof("notify module config")

	err := config.NotifyModuleConfigForHttp(ctx, nconfig)
	if err != nil {
		ctxlog.Errorf("notify module config err:%+v", err)
	}

	SendResponse(c, &agent.AgentctlResponse{
		Successful: true,
		Message:    "notify module config success",
		Error:      "",
	}, err)
}

func ValidateConfigPropertiesHandler(c *gin.Context) {
	kvs := config.KeyValues{}
	c.Bind(&kvs)

	ctx := NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx)
	ctxlog.Debugf("validate module config")

	err := config.ValidateConfigKeyValues(ctx, kvs.Configs)
	if err != nil {
		ctxlog.Errorf("validate configs failed, err:%+v", err)
	}

	SendResponse(c, &agent.AgentctlResponse{
		Successful: true,
		Message:    "success",
		Error:      "",
	}, err)
}

// http handler: effect module config
func ConfigStatusHandler(c *gin.Context) {
	needRestartModuleKeyValues := config.NeedRestartModuleKeyValues()
	SendResponse(c, needRestartModuleKeyValues, nil)
}

func ReloadConfigHandler(c *gin.Context) {
	ctx := NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx)
	err := mgragent.GlobalConfigManager.ReloadModuleConfigs(ctx)
	if err != nil {
		ctxlog.Errorf("reload module config files failed, err: %+v", err)
		SendResponse(c, nil, err)
		return
	}
	err = config.NotifyAllModules(ctx)
	if err != nil {
		ctxlog.Errorf("notify config change after changing module config failed, err: %+v", err)
		SendResponse(c, nil, err)
		return
	}
	SendResponse(c, "reload config success", nil)
}

func ChangeConfigHandler(c *gin.Context) {
	req := mgragent.ModuleConfigChangeRequest{
		Reload: true,
	}
	ctx := NewContextWithTraceId(c)
	ctxlog := log.WithContext(ctx)

	err := c.BindJSON(&req)
	if err != nil {
		ctxlog.Errorf("parse request failed, err: %+v", err)
		SendResponse(c, nil, err)
		return
	}
	ctxlog.Debugf("change module config %+v", req)
	changedFileNames, err := mgragent.GlobalConfigManager.ChangeModuleConfigs(ctx, &req)
	if err != nil {
		ctxlog.Errorf("change module config failed, err: %+v", err)
		SendResponse(c, nil, err)
		return
	}
	ctxlog.Infof("module config files changed: %+v", changedFileNames)
	if len(changedFileNames) > 0 && req.Reload {
		err = mgragent.GlobalConfigManager.ReloadModuleConfigs(ctx)
		if err != nil {
			ctxlog.Errorf("reload module config files failed, err: %+v", err)
			SendResponse(c, nil, err)
			return
		}
		err = config.NotifyAllModules(ctx)
		if err != nil {
			ctxlog.Errorf("notify config change after changing module config failed, err: %+v", err)
			SendResponse(c, nil, err)
			return
		}
	}
	SendResponse(c, fmt.Sprintf("%d files changed", len(req.ModuleConfigChanges)), nil)
}
