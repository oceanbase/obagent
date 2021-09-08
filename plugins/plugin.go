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

package plugins

import (
	"bytes"

	"github.com/oceanbase/obagent/metric"
)

//Initializer contains Init function
type Initializer interface {
	Init(config map[string]interface{}) error
}

//Closer contains Close function
type Closer interface {
	Close() error
}

//Describer show SampleConfig and Description
type Describer interface {
	SampleConfig() string
	Description() string
}

//Module contains Initializer Closer Describer
type Module interface {
	Initializer
	Closer
	Describer
}

//Input used to collect metrics
type Input interface {
	Module
	Collect() ([]metric.Metric, error)
}

//Processor used to process metrics
type Processor interface {
	Module
	Process(metrics ...metric.Metric) ([]metric.Metric, error)
}

//Output write metrics to target in push mode
type Output interface {
	Module
	Write(metrics []metric.Metric) error
}

//Exporter convert metrics to buffer and export in pull mode
type Exporter interface {
	Module
	Export(metrics []metric.Metric) (*bytes.Buffer, error)
}
