// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package engine

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/config"
)

func TestMonitorAgentServerShutdown(t *testing.T) {
	server := NewMonitorAgentServer(&config.MonitorAgentConfig{
		Log: &config.LogConfig{Level: "debug"},
		Server: config.MonitorAgentHttpConfig{
			Address:      ":62889",
			AdminAddress: ":62886",
		},
	})
	go server.Run()

	t.Run("shutdown without any request", func(t *testing.T) {
		err := server.Server.Shutdown(context.Background())
		Convey("shutdown err", t, func() {
			So(err, ShouldBeNil)
		})

		adminerr := server.AdminServer.Shutdown(context.Background())
		Convey("shutdown adminerr", t, func() {
			So(adminerr, ShouldBeNil)
		})
	})

}
