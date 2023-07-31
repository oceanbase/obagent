package web

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/config/monagent"
)

func TestMonitorAgentServerShutdown(t *testing.T) {
	server := NewMonitorAgentServer(&monagent.MonitorAgentConfig{
		Server: monagent.MonitorAgentHttpConfig{
			Address: ":62889",
		},
	})
	go server.Run()

	t.Run("shutdown without any request", func(t *testing.T) {
		err := server.Server.Shutdown(context.Background())
		Convey("shutdown err", t, func() {
			So(err, ShouldBeNil)
		})
	})

}
