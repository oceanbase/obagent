package mgragent

import (
	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/executor/file"
)

func isFileExists(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	var param file.GetFileExistsParam
	c.BindJSON(&param)
	data, err := file.IsFileExists(ctx, param)
	common.SendResponse(c, data, err)
}

func getRealStaticPath(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	var param file.GetRealStaticPathParam
	c.BindJSON(&param)
	data, err := file.GetRealStaticPath(ctx, param)
	common.SendResponse(c, data, err)
}
