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
