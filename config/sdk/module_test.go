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

package sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/config"
)

func TestLogModule(t *testing.T) {
	err := initSDK()
	assert.Nil(t, err)

	ctx := context.Background()
	// init log
	err = config.InitModuleConfig(ctx, config.ManagerLogConfigModule)
	assert.Nil(t, err)

	// init basic auth
	common.InitBasicAuthConf(ctx)

	// init notify process
	err = config.InitModuleConfig(ctx, config.NotifyProcessConfigModule)
	assert.Nil(t, err)

	err = config.InitModuleConfig(ctx, config.OBLogcleanerModule)
	assert.Nil(t, err)

}
