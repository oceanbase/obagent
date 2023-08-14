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

package mgragent

import (
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/stat"
)

func InitManagerAgentRoutes(s *http.StateHolder, r *gin.Engine) {
	r.Use(common.HttpStatMiddleware)

	// self stat metrics
	r.GET("/metrics/stat", adapter.Wrap(stat.PromHandler))
	r.Use(
		gin.CustomRecovery(common.Recovery), // gin's crash-free middleware
		common.PreHandlers("/api/v1/module/config/update", "/api/v1/module/config/validate"),
		common.SetContentType,
		common.PostHandlers("/debug/pprof"),
	)

	v1 := r.Group("/api/v1")
	v1.GET("/time", common.TimeHandler)
	v1.GET("/info", common.InfoHandler)
	v1.GET("/git-info", common.GitInfoHandler)
	v1.GET("/status", common.StatusHandler(s))
	v1.POST("/status", common.StatusHandler(s))

	// task routes
	task := v1.Group("/task")
	task.POST("/status", queryTaskHandler)
	task.GET("/status", queryTaskHandler)

	// agent admin routes
	agent := v1.Group("/agent")
	agent.POST("/status", agentStatusService)
	agent.GET("/status", agentStatusService)
	agent.POST("/restart", asyncCommandHandler(restartCmd))

	// file routes
	file := v1.Group("/file")
	file.POST("/exists", isFileExists)
	file.POST("/getRealPath", getRealStaticPath)

	// system routes
	system := v1.Group("/system")
	system.POST("/hostInfo", getHostInfoHandler)

	// module config
	v1.POST("/module/config/update", common.UpdateConfigPropertiesHandler)
	v1.POST("/module/config/notify", common.NotifyConfigPropertiesHandler)
	v1.POST("/module/config/validate", common.ValidateConfigPropertiesHandler)
	v1.GET("/module/config/status", common.ConfigStatusHandler)
	v1.POST("/module/config/change", common.ChangeConfigHandler)
	v1.POST("/module/config/reload", common.ReloadConfigHandler)

	logGroup := v1.Group("/log")
	logGroup.POST("/query", queryLogHandler)
	logGroup.POST("/download", downloadLogHandler)

	r.NoRoute(func(c *gin.Context) {
		err := errors.Occur(errors.ErrBadRequest, "404 not found")
		common.SendResponse(c, nil, err)
	})
}
