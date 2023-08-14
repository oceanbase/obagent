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
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/lib/command"
)

var restartCmd = command.WrapFunc(func(taskToken agent.TaskToken) error {
	log.Info("restarting agent")
	cmd := agent.NewAgentctlCmd()
	return cmd.Restart(taskToken)
})

func agentStatusService(c *gin.Context) {
	log.Info("query agent status")
	admin := agent.NewAdmin(agent.DefaultAdminConf())
	status, err := admin.AgentStatus()
	if err != nil {
		common.SendResponse(c, nil, err)
		return
	}
	common.SendResponse(c, status, nil)
}
