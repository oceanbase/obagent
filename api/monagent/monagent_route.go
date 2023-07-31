package monagent

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/config"
	http2 "github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/lib/system"
	"github.com/oceanbase/obagent/stat"
)

func InitMonitorAgentRoutes(router *gin.Engine, localRouter *gin.Engine) {
	router.GET("/metrics/stat", adapter.Wrap(stat.PromHandler))

	v1 := router.Group("/api/v1")
	v1.Use(common.PostHandlers())

	v1.POST("/module/config/update", common.UpdateConfigPropertiesHandler)
	v1.POST("/module/config/notify", common.NotifyConfigPropertiesHandler)
	v1.POST("/module/config/validate", common.ValidateConfigPropertiesHandler)

	v1.GET("/time", common.TimeHandler)
	v1.GET("/info", common.InfoHandler)
	v1.GET("/git-info", common.GitInfoHandler)
	v1.POST("/status", monitorStatusHandler)
	v1.GET("/status", monitorStatusHandler)

	initMonagentLocalRoutes(localRouter)
}

func initMonagentLocalRoutes(localRouter *gin.Engine) {
	common.InitPprofRouter(localRouter)

	localRouter.GET("/metrics/stat", adapter.Wrap(stat.PromHandler))

	group := localRouter.Group("/api/v1")
	group.GET("/time", common.TimeHandler)
	group.GET("/info", common.InfoHandler)
	group.POST("/info", common.InfoHandler)
	group.GET("/git-info", common.GitInfoHandler)
	group.POST("/status", monitorStatusHandler)
	group.GET("/status", monitorStatusHandler)
	group.POST("/module/config/update", common.UpdateConfigPropertiesHandler)
	group.POST("/module/config/notify", common.NotifyConfigPropertiesHandler)
}

func UseLocalMonitorMiddleware(r *gin.Engine) {
	r.Use(
		common.HttpStatMiddleware,
		gin.CustomRecovery(common.Recovery), // gin's crash-free middleware
		common.PreHandlers("/api/v1/module/config/update", "/api/v1/module/config/validate"),
		common.PostHandlers("/debug/pprof", "/debug/fgprof", "/metrics/", "/api/v1/log/alarms"),
	)
}

func UseMonitorMiddleware(r *gin.Engine) {
	r.Use(
		common.HttpStatMiddleware,
		gin.CustomRecovery(common.Recovery), // gin's crash-free middleware
		common.PreHandlers("/api/v1/module/config/update", "/api/v1/module/config/validate"),
		common.MonitorAgentPostHandler,
	)
}

func RegisterPipelineRoute(ctx context.Context, r *gin.Engine, url string, fh func(http.Handler) http.Handler) {
	log.WithContext(ctx).Infof("register route %s", url)
	r.GET(url, adapter.Wrap(fh))
}

var libProcess system.Process = system.ProcessImpl{}

func monitorStatusHandler(c *gin.Context) {
	ports := make([]int, 0)

	pid := os.Getpid()
	processInfo, err := libProcess.GetProcessInfoByPid(int32(pid))
	if err != nil {
		log.Errorf("StatusHandler get processInfo failed, pid:%s", pid)
	} else {
		ports = processInfo.Ports
	}
	var info = http2.Status{
		State:   http2.Running,
		Version: config.AgentVersion,
		Pid:     pid,
		StartAt: common.StartAt,
		Ports:   ports,
	}
	common.SendResponse(c, info, nil)
}
