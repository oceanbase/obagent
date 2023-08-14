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

package system

import (
	"context"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/system"
)

var libSystem system.System = system.SystemImpl{}

func GetHostInfo(ctx context.Context) (*system.HostInfo, *errors.OcpAgentError) {
	info, err := libSystem.GetHostInfo()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return info, nil
}
