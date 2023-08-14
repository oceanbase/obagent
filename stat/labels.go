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

package stat

const PluginTypeKey = "type"

const PluginExecuteStatusKey = "status"

const PluginNameKey = "plugin"

const (
	MysqlOutputMetricName   = "metric_name"
	MysqlOutputTableNameKey = "table"
	MysqlOutputTaskNameKey  = "task_name"
)

const (
	Process     = "process"
	HttpMethod  = "method"
	HttpStatus  = "status"
	HttpApiPath = "path"
	SvrIP       = "svr_ip"
	App         = "app"
	Host        = "HOST"
)

var HostIP string

const (
	LogFileName = "log_file_name"
)
