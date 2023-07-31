package agentd

import (
	"testing"

	"github.com/oceanbase/obagent/lib/path"
)

func Test_serviceProc(t *testing.T) {
	conf := toProcConfig(ServiceConfig{})
	if conf.Cwd != path.AgentDir() {
		t.Error("default cwd wrong")
	}
	if !conf.InheritEnv {
		t.Error("InheritEnv should be true")
	}
}
