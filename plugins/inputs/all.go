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

package inputs

import (
	"github.com/oceanbase/obagent/plugins"
	"github.com/oceanbase/obagent/plugins/inputs/mysql"
	"github.com/oceanbase/obagent/plugins/inputs/nodeexporter"
	"github.com/oceanbase/obagent/plugins/inputs/oceanbase/log"
	"github.com/oceanbase/obagent/plugins/inputs/prometheus"
)

func init() {
	plugins.GetInputManager().Register("mysqlTableInput", func() plugins.Input {
		return &mysql.TableInput{}
	})
	plugins.GetInputManager().Register("mysqldInput", func() plugins.Input {
		return &mysql.MysqldExporter{}
	})
	plugins.GetInputManager().Register("nodeExporterInput", func() plugins.Input {
		return &nodeexporter.NodeExporter{}
	})
	plugins.GetInputManager().Register("prometheusInput", func() plugins.Input {
		return &prometheus.Prometheus{}
	})
	plugins.GetInputManager().Register("errorLogInput", func() plugins.Input {
		return &log.ErrorLogInput{}
	})
}
