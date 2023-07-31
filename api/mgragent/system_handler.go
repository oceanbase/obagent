package mgragent

import (
	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/executor/system"
)

func getHostInfoHandler(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	data, err := system.GetHostInfo(ctx)
	common.SendResponse(c, data, err)
}
