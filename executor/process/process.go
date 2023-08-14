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

package process

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/shell"
	"github.com/oceanbase/obagent/lib/system"
)

var libShell shell.Shell = shell.ShellImpl{}

type FindProcessType string

const (
	byName    FindProcessType = "BY_NAME"
	byKeyword FindProcessType = "BY_KEYWORD"
	byPid     FindProcessType = "BY_PID"
)

type FindProcessParam struct {
	FindType FindProcessType `json:"findType" binding:"required,oneof='BY_NAME' 'BY_KEYWORD'"` // find process by name or keyword
	Name     string          `json:"name" binding:"required_if=FindType BY_NAME"`              // process name
	Keyword  string          `json:"keyword" binding:"required_if=FindType BY_KEYWORD"`        // process keyword
}

type CheckProcessExistsParam struct {
	Name string `json:"name" binding:"required"` // process name
}

type GetProcessProcInfoParam struct {
	FindType FindProcessType `json:"findType" binding:"required,oneof='BY_NAME' 'BY_PID'"`
	Pid      int32           `json:"pid"  binding:"required_if=FindType BY_PID"`  // pid
	Name     string          `json:"name" binding:"required_if=FindType BY_NAME"` // process name
	Type     string          `json:"type" binding:"required"`                     // proc type
	User     string          `json:"user" binding:"required"`                     // user
}

type GetProcessInfoParam struct {
	Name string `json:"processName" binding:"required"` // process name
}

type StopProcessParam struct {
	Process FindProcessParam `json:"process" binding:"required"` // process name or process keyword
	Force   bool             `json:"force"`                      // whether to force stop process
}

type GetProcessInfoResult struct {
	ProcessInfoList []*system.ProcessInfo `json:"processInfoList"`
}

type GetProcessProcInfoResult struct {
	ProcInfoList []string `json:"procInfoList"`
}

var libProcess system.Process = system.ProcessImpl{}

func ProcessExists(ctx context.Context, param CheckProcessExistsParam) (bool, *errors.OcpAgentError) {
	name := param.Name
	ctxlog := log.WithContext(ctx).WithField("name", name)

	exists, err := libProcess.ProcessExists(name)
	if err != nil {
		return false, errors.Occur(errors.ErrCheckProcessExists, name, err)
	}
	ctxlog.WithField("exists", exists).Info("check process exists done")
	return exists, nil
}

func GetProcessInfo(ctx context.Context, param GetProcessInfoParam) (*GetProcessInfoResult, *errors.OcpAgentError) {
	name := param.Name
	ctxlog := log.WithContext(ctx).WithField("name", name)

	processInfoList, err := libProcess.FindProcessInfoByName(name)
	if err != nil {
		return nil, errors.Occur(errors.ErrGetProcessInfo, name, err)
	}
	ctxlog.WithField("processes", processInfoList).Info("get process info done")
	return &GetProcessInfoResult{ProcessInfoList: processInfoList}, nil
}

func StopProcess(ctx context.Context, param StopProcessParam) *errors.OcpAgentError {
	force := param.Force

	if param.Process.FindType == byName {
		name := param.Process.Name
		if force {
			err := libProcess.KillProcessByName(name)
			if err != nil {
				return errors.Occur(errors.ErrStopProcess, name, err)
			}
		} else {
			err := libProcess.TerminateProcessByName(name)
			if err != nil {
				return errors.Occur(errors.ErrStopProcess, name, err)
			}
		}
	} else if param.Process.FindType == byKeyword {
		keyword := param.Process.Keyword
		if force {
			err := libProcess.KillProcessByKeyword(keyword)
			if err != nil {
				return errors.Occur(errors.ErrStopProcess, keyword, err)
			}
		} else {
			err := libProcess.TerminateProcessByKeyword(keyword)
			if err != nil {
				return errors.Occur(errors.ErrStopProcess, keyword, err)
			}
		}
	}
	log.WithContext(ctx).WithFields(log.Fields{
		"findType": param.Process.FindType,
		"name":     param.Process.Name,
		"keyword":  param.Process.Keyword,
		"force":    force,
	}).Info("stop process done")
	return nil
}

// GetProcessProcInfo ls -l /proc/{pid}/{infoType}
// even root user can not list this info of others' in container, so user must be specified
func GetProcessProcInfo(ctx context.Context, param GetProcessProcInfoParam) (*GetProcessProcInfoResult, *errors.OcpAgentError) {
	user := param.User
	infoType := param.Type
	var lines []string
	var err error
	var processIdentifier string
	if param.FindType == byPid {
		pid := param.Pid
		processIdentifier = string(pid)
		log.WithContext(ctx).WithFields(log.Fields{
			"pid":      pid,
			"user":     user,
			"infoType": infoType,
		}).Info("list process proc info")
		lines, err = libProcess.GetProcessProcInfoByPid(pid, infoType, user)
	} else if param.FindType == byName {
		name := param.Name
		processIdentifier = name
		log.WithContext(ctx).WithFields(log.Fields{
			"name":     name,
			"user":     user,
			"infoType": infoType,
		}).Info("list process proc info")
		lines, err = libProcess.GetProcessProcInfoByName(name, infoType, user)
	}
	if err != nil {
		return nil, errors.Occur(errors.ErrProcessProcInfo, infoType, processIdentifier, user, err)
	}
	return &GetProcessProcInfoResult{ProcInfoList: lines}, nil
}
