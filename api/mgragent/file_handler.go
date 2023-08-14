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
