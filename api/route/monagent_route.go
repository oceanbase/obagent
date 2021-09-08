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
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/stat"
)

func InitMonagentRoutes(r *gin.Engine) {
	r.Use(
		gin.Recovery(), // gin's crash-free middleware
	)

	v1 := r.Group("/api/v1")
	v1.POST("/module/config/update", updateModuleConfigHandler)
	v1.POST("/module/config/notify", notifyModuleConfigHandler)

	metric := r.Group("/metrics")
	metric.GET("/stat", adapter.Wrap(stat.PromGinWrapper))
}

func RegisterPipelineRoute(r *gin.Engine, url string, fh func(http.Handler) http.Handler) {
	log.Infof("register route %s", url)
	r.GET(url, adapter.Wrap(fh))
}

func InitPprofRouter(r *gin.Engine) {
	pprof.Register(r, "debug/pprof")
}
