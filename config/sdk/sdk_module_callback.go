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

package sdk

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	"github.com/oceanbase/obagent/config"
)

//RegisterConfigCallbacks To load the business callback module,
//the business needs to provide a callback function,
//and the configuration can be obtained when the business is initialized or changed.
func RegisterConfigCallbacks() error {
	// register config module callbacks
	err := config.RegisterConfigCallback(
		config.NotifyProcessConfigModuleType,
		func() interface{} {
			return []config.ProcessConfigNotifyAddress{}
		},
		// Initial configuration callback
		processNotifyModuleConfigCallback,
		// Configuration update callback
		processNotifyModuleConfigCallback,
	)
	if err != nil {
		return err
	}

	return nil
}

func processNotifyModuleConfigCallback(ctx context.Context, moduleConf interface{}) error {
	confs, ok := moduleConf.([]config.ProcessConfigNotifyAddress)
	if !ok {
		return errors.Errorf("module %s conf %s is not NotifyAddress", config.NotifyProcessConfigModuleType, reflect.TypeOf(moduleConf))
	}
	for _, conf := range confs {
		config.SetProcessModuleConfigNotifyAddress(conf)
	}
	return nil
}
