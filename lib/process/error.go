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

package process

import "github.com/oceanbase/obagent/lib/errors"

var (
	FinalWaitTimeoutErr = errors.DeadlineExceeded.NewCode("lib/process", "final_wait_timeout").
				WithMessageTemplate("process not killed after final wait %v")
	ProcNotExistsErr = errors.NotFound.NewCode("lib/process", "proc_not_exists").
				WithMessageTemplate("process not exists")
	ProcAlreadyRunningErr = errors.FailedPrecondition.NewCode("lib/process", "proc_already_running").
				WithMessageTemplate("process already running")
	PidFileExistsErr = errors.FailedPrecondition.NewCode("lib/process", "pid_file_exists").
				WithMessageTemplate("pid file '%s' exists")
	PidFileNotOpenedErr = errors.FailedPrecondition.NewCode("lib/process", "pid_file_not_opened").
				WithMessageTemplate("pid file not opened")
	WritePidFailedErr     = errors.Internal.NewCode("lib/process", "write_pid_failed")
	FailToStartProcessErr = errors.Internal.NewCode("lib/process", "fail_to_start_process")
	PipeOutputErr         = errors.Internal.NewCode("lib/process", "failed_to_pipe_stdout_stderr")
	SignalErr             = errors.Internal.NewCode("lib/process", "failed_send_signal").
				WithMessageTemplate("failed to send signal %s to %d")
	CreateCmdFailErr = errors.Internal.NewCode("lib/process", "create_cmd_fail")
	InvalidRlimitErr = errors.InvalidArgument.NewCode("lib/process", "invalid_rlimit_name")
	UserGroupErr     = errors.InvalidArgument.NewCode("lib/process", "user_group_err")
)
