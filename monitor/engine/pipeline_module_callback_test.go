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
