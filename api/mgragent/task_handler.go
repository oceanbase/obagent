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

package mgragent

import (
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/lib/command"
	path2 "github.com/oceanbase/obagent/lib/path"
)

type QueryTaskParam struct {
	TaskToken string `json:"taskToken"`
}

type TaskTokenResult struct {
	TaskToken string `json:"taskToken"`
}

type TaskStatusResult struct {
	Finished bool        `json:"finished"`
	Ok       bool        `json:"ok"`
	Result   interface{} `json:"result"`
	Err      string      `json:"err"`
	Progress interface{} `json:"progress"`
}

var taskExecutor = command.NewExecutor(command.NewFileTaskStore(path2.TaskStoreDir()))

func queryTaskHandler(c *gin.Context) {
	//ctx := NewContextWithTraceId(c)
	var param QueryTaskParam
	c.BindJSON(&param)
	status, ok := taskExecutor.GetResult(command.ExecutionTokenFromString(param.TaskToken))
	if !ok {
		common.SendResponse(c, nil, errors.Occur(errors.ErrTaskNotFound, param.TaskToken))
		return
	}
	common.SendResponse(c, TaskStatusResult{
		Finished: status.Finished,
		Ok:       status.Ok,
		Result:   status.Result,
		Err:      status.Err,
		Progress: status.Progress,
	}, nil)
}

func TaskCount() int {
	return len(taskExecutor.AllExecutions())
}

func asyncCommandHandler(task command.Command) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := common.NewContextWithTraceId(c)
		defaultParam := task.DefaultParam()
		v := reflect.New(reflect.TypeOf(defaultParam))
		v.Elem().Set(reflect.ValueOf(defaultParam))
		param := v.Interface()
		err := c.BindJSON(param)
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		taskToken := param.(agent.TaskTokenParam)
		if taskToken.GetTaskToken() == "" {
			taskToken.SetTaskToken(command.GenerateTaskId())
		}
		input := command.NewInput(ctx, reflect.ValueOf(param).Elem().Interface())
		input.WithRequestTaskToken(taskToken.GetTaskToken())
		token, err := taskExecutor.Execute(task, input)
		if err != nil {
			common.SendResponse(c, nil, err)
			return
		}
		common.SendResponse(c, TaskTokenResult{
			TaskToken: token.String(),
		}, nil)
	}
}
