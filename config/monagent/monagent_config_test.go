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

package monagent

import (
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDecodeMonitorAgentServerConfig(t *testing.T) {
	t.Run("DecodeMonitorAgentServerConfig success", func(t *testing.T) {
		serverConfig, err := DecodeMonitorAgentServerConfig("../../etc/monagent.yaml")
		Convey("DecodeMonitorAgentServerConfig success", t, func() {
			So(err, ShouldBeNil)
			So(serverConfig, ShouldNotBeNil)
			So(serverConfig.Server.Address, ShouldEqual, "0.0.0.0:${ocp.agent.monitor.http.port}")
		})
	})

	t.Run("DecodeMonitorAgentServerConfig with no such file", func(t *testing.T) {
		go func() {
			if err := recover(); err != nil {
				logrus.Info(err)
			}
		}()
		serverConfig, err := DecodeMonitorAgentServerConfig("/do/not/exist/config.yaml")
		logrus.Info(err)
		Convey("DecodeMonitorAgentServerConfig", t, func() {
			So(serverConfig, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "no such file or director")
		})
	})

}
