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

package sdk

import (
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/lib/path"
)

// Set the configuration item meta information.
// All configuration items have to be configured with meta information: otherwise,
// configuration items cannot be parsed according to their data type.
func setConfigPropertyMeta() {
	setLogConfigPropertyMeta()
	setLogCleanerConfigPropertyMeta()
	setLogtailerConfigPropertyMeta()
	setMonitorAgentConfigPropertyMeta()
	setBasicAuthConfigPropertyMeta()
	setCommonAgentConfigPropertyMeta()
}

func setBasicAuthConfigPropertyMeta() {
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "agent.http.basic.auth.username",
			DefaultValue: "ocp_agent",
			ValueType:    config.ValueString,
			Fatal:        false,
			Masked:       false,
			NeedRestart:  false,
			Description:  "basic auth username",
			Unit:         "",
			Valid:        nil,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "agent.http.basic.auth.password",
			DefaultValue: "",
			ValueType:    config.ValueString,
			Encrypted:    true,
			Fatal:        false,
			Masked:       true,
			NeedRestart:  false,
			Description:  "basic auth password",
			Unit:         "",
			Valid:        nil,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "agent.http.basic.auth.metricAuthEnabled",
			DefaultValue: "true",
			ValueType:    config.ValueBool,
			Encrypted:    false,
			Fatal:        false,
			Masked:       false,
			NeedRestart:  true,
			Description:  "basic auth disabled",
			Unit:         "",
			Valid:        nil,
		})
}

func setCommonAgentConfigPropertyMeta() {
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "obagent.home.path",
			DefaultValue: path.AgentDir(),
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ocp.agent.home.path",
			DefaultValue: path.AgentDir(),
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ocp.agent.http.socks.proxy.enabled",
			DefaultValue: "false",
			ValueType:    config.ValueBool,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ocp.agent.http.socks.proxy.address",
			DefaultValue: "",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ocp.agent.manager.http.port",
			DefaultValue: 62888,
			ValueType:    config.ValueInt64,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ocp.agent.monitor.http.port",
			DefaultValue: 62889,
			ValueType:    config.ValueInt64,
		})
}
