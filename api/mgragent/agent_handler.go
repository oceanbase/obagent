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
