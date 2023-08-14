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

package agent

import (
	"strings"
	"testing"

	"github.com/oceanbase/obagent/agentd/api"
)

func TestAgentctlCmd(t *testing.T) {
	agentctl := NewAgentctlCmd()
	if !strings.HasSuffix(agentctl.agentctlPath, "/ob_agentctl") {
		t.Error("bad agentctl path")
	}
	agentctl.agentctlPath = "echo"
	err := agentctl.Restart(TaskToken{
		TaskToken: "test-token",
	})
	if err != nil {
		t.Error(err)
	}
	err = agentctl.Reinstall(ReinstallParam{})
	if err != nil {
		t.Error(err)
	}

	err = agentctl.StartService(StartStopServiceParam{
		TaskToken: TaskToken{"test-token"},
		StartStopAgentParam: api.StartStopAgentParam{
			Service: "ob_mgragent",
		},
	})
	if err != nil {
		t.Error(err)
	}

	err = agentctl.StopService(StartStopServiceParam{
		TaskToken: TaskToken{"test-token"},
		StartStopAgentParam: api.StartStopAgentParam{
			Service: "ob_mgragent",
		},
	})
	if err != nil {
		t.Error(err)
	}
}
