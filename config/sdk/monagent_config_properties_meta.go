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

func setMonitorAgentConfigPropertyMeta() {

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.ob.monitor.user",
			DefaultValue: "ocp_monitor",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.ob.monitor.password",
			DefaultValue: "",
			ValueType:    config.ValueString,
			Encrypted:    true,
			Masked:       true,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.ob.sql.port",
			DefaultValue: "2881",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.ob.rpc.port",
			DefaultValue: "2882",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.host.ip",
			DefaultValue: "127.0.0.1",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.cluster.id",
			DefaultValue: "0",
			ValueType:    config.ValueInt64,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.ob.cluster.name",
			DefaultValue: "",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.ob.cluster.id",
			DefaultValue: "0",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.ob.zone.name",
			DefaultValue: "",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.pipeline.ob.status",
			DefaultValue: "inactive",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.pipeline.node.status",
			DefaultValue: "active",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.log.path",
			DefaultValue: "/data/log1",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.data.path",
			DefaultValue: "/data/1",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.install.path",
			DefaultValue: "/home/admin/oceanbase",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "host.check.readonly.mountpoint",
			DefaultValue: "/",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.node.custom.interval",
			DefaultValue: "1s",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.pipeline.ob.log.status",
			DefaultValue: "inactive",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "es.client.addresses",
			DefaultValue: "",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "es.client.auth.username",
			DefaultValue: "",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "es.client.auth.password",
			DefaultValue: "",
			ValueType:    config.ValueString,
			Encrypted:    true,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "observer.log.path",
			DefaultValue: "/home/admin/oceanbase/log",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "agent.log.path",
			DefaultValue: path.LogDir(),
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "os.log.path",
			DefaultValue: "/var/log",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.second.metric.cache.update.interval",
			DefaultValue: "15s",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.collector.prometheus.interval",
			DefaultValue: "15s",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.collector.ob.basic.interval",
			DefaultValue: "15s",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.collector.ob.extra.interval",
			DefaultValue: "60s",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.collector.ob.snapshot.interval",
			DefaultValue: "1h",
			ValueType:    config.ValueString,
		})

	// obagent 1.2.0
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.mysql.monitor.user",
			DefaultValue: "mysql_monitor_user",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.mysql.monitor.password",
			DefaultValue: "mysql_monitor_password",
			ValueType:    config.ValueString,
			Encrypted:    true,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.mysql.sql.port",
			DefaultValue: 3306,
			ValueType:    config.ValueInt64,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.mysql.host",
			DefaultValue: "127.0.0.1",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.pipeline.mysql.status",
			DefaultValue: "inactive",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.alertmanager.address",
			DefaultValue: "",
			ValueType:    config.ValueString,
		})
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.pipeline.ob.alertmanager.status",
			DefaultValue: "inactive",
			ValueType:    config.ValueString,
		})
}
