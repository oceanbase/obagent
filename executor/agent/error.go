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

package agent

import "github.com/oceanbase/obagent/lib/errors"

var (
	ReadPidFailedErr                = errors.Internal.NewCode("agent", "read_pid_failed")
	BackupPidFailedErr              = errors.Internal.NewCode("agent", "backup_pid_failed")
	RestorePidFailedErr             = errors.Internal.NewCode("agent", "restore_pid_failed")
	ExecuteAgentctlFailedErr        = errors.Internal.NewCode("agent", "execute_agentctl_failed")
	FetchAdminLockFailedErr         = errors.Internal.NewCode("agent", "fetch_admin_lock_failed")
	CleanDanglingAdminLockFailedErr = errors.Internal.NewCode("agent", "clean_dangling_admin_lock_failed")
	ReleaseAdminLockFailedErr       = errors.Internal.NewCode("agent", "release_admin_lock_failed")
	ChecksumNotMatchErr             = errors.FailedPrecondition.NewCode("agent", "checksum_not_match")
	AgentdNotRunningErr             = errors.FailedPrecondition.NewCode("agent", "agentd_not_running")
	AgentdAlreadyRunningErr         = errors.FailedPrecondition.NewCode("agent", "agentd_already_running")
	WaitForReadyTimeoutErr          = errors.DeadlineExceeded.NewCode("agent", "wait_for_ready_timeout")
	WaitForExitTimeoutErr           = errors.DeadlineExceeded.NewCode("agent", "wait_for_exit_timeout")
	AgentdExitedQuicklyErr          = errors.Internal.NewCode("agent", "agentd_exited_quickly")

	AgentctlStopServiceFailedErr  = errors.Internal.NewCode("agentctl", "agentctl_stop_service_failed")
	AgentctlStartServiceFailedErr = errors.Internal.NewCode("agentctl", "agentctl_start_service_failed")
	AgentctlRestartFailedErr      = errors.Internal.NewCode("agentctl", "agentctl_restart_failed")
	AgentctlReinstallFailedErr    = errors.Internal.NewCode("agentctl", "agentctl_reinstall_failed")
)
