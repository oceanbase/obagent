/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package agentd

import "github.com/oceanbase/obagent/lib/errors"

const module = "agentd"

var (
	ServiceAlreadyStartedErr = errors.FailedPrecondition.NewCode(module, "service_already_started")
	ServiceAlreadyStoppedErr = errors.FailedPrecondition.NewCode(module, "service_already_stopped")
	ServiceNotFoundErr       = errors.NotFound.NewCode(module, "service_not_round")
	BadParamErr              = errors.InvalidArgument.NewCode(module, "bad_param").WithMessageTemplate("invalid input parameter, maybe bad format: %v")
	InvalidActionErr         = errors.InvalidArgument.NewCode(module, "invalid_action").WithMessageTemplate("invalid action: %s")
	InternalServiceErr       = errors.Internal.NewCode(module, "internal_service_err")
	AgentdNotRunningErr      = errors.Internal.NewCode(module, "agentd_not_running")
	WritePidFailedErr        = errors.Internal.NewCode(module, "write_pid_failed")
	RemovePidFailedErr       = errors.Internal.NewCode(module, "remove_pid_failed")
)
