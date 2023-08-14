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

package engine

import (
	"context"
	"testing"

	"github.com/oceanbase/obagent/config/monagent"
)

func TestAddPipeline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go GetPipelineManager().emptySchedule(ctx)
	go GetConfigManager().Schedule(ctx)

	testPipelineModule := &monagent.PipelineModule{}
	err := addPipeline(ctx, testPipelineModule)
	if err != nil {
		t.Errorf("test add pipeline failed %s", err.Error())
	}

	_ = addPipeline(ctx, testPipelineModule)

}

func TestDeletePipeline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go GetPipelineManager().emptySchedule(ctx)
	go GetConfigManager().Schedule(ctx)

	testPipelineModule := &monagent.PipelineModule{}
	err := deletePipeline(ctx, testPipelineModule)
	if err != nil {
		t.Errorf("test delete pipeline failed %s", err.Error())
	}

	_ = deletePipeline(ctx, testPipelineModule)

}

func TestUpdatePipeline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go GetPipelineManager().emptySchedule(ctx)
	go GetConfigManager().Schedule(ctx)

	testPipelineModule := &monagent.PipelineModule{}
	err := addPipeline(ctx, testPipelineModule)
	if err != nil {
		t.Errorf("test update pipeline failed %s", err.Error())
	}

	_ = updatePipeline(ctx, testPipelineModule)
}
