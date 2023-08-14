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

package agent

import (
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/mask"
	"github.com/oceanbase/obagent/lib/path"
)

type AgentctlCmd struct {
	agentctlPath string
}

// AgentctlResponse the formatted content written to stdout or stderr
type AgentctlResponse struct {
	Successful bool        `json:"successful"`
	Message    interface{} `json:"message"`
	Error      string      `json:"error"`
}

func NewAgentctlCmd() *AgentctlCmd {
	return &AgentctlCmd{
		agentctlPath: path.AgentctlPath(),
	}
}

func (c *AgentctlCmd) StopService(param StartStopServiceParam) error {
	args := []string{"service", "stop", param.Service, "--task-token=" + param.TaskToken.TaskToken}
	log.Infof("execute agentctl '%s' %v", c.agentctlPath, args)
	result, err := exec.Command(c.agentctlPath, args...).CombinedOutput()
	if err != nil {
		return AgentctlStopServiceFailedErr.NewError(string(result)).WithCause(err)
	}
	return nil
}

func (c *AgentctlCmd) StartService(param StartStopServiceParam) error {
	args := []string{"service", "start", param.Service, "--task-token=" + param.TaskToken.TaskToken}
	log.Infof("execute agentctl '%s' %v", c.agentctlPath, args)
	result, err := exec.Command(c.agentctlPath, args...).CombinedOutput()
	if err != nil {
		return AgentctlStartServiceFailedErr.NewError(string(result)).WithCause(err)
	}
	return nil
}

func (c *AgentctlCmd) Restart(token TaskToken) error {
	args := []string{"restart", "--task-token=" + token.TaskToken}
	log.Infof("execute agentctl '%s' %v", c.agentctlPath, args)
	result, err := exec.Command(c.agentctlPath, args...).CombinedOutput()
	if err != nil {
		return AgentctlRestartFailedErr.NewError(string(result)).WithCause(err)
	}
	return nil
}

func (c *AgentctlCmd) Reinstall(param ReinstallParam) error {
	args := []string{
		"reinstall",
		"--source=" + param.Source,
		"--checksum=" + param.Checksum,
		"--version=" + param.Version,
		"--task-token=" + param.TaskToken.TaskToken,
	}
	log.Infof("execute agentctl '%s' %v", c.agentctlPath, mask.MaskSlice(args))
	result, err := exec.Command(c.agentctlPath, args...).CombinedOutput()
	if err != nil {
		return AgentctlReinstallFailedErr.NewError(string(result)).WithCause(err)
	}
	return nil
}
