package engine

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/oceanbase/obagent/config/monagent"
)

func TestCreateModule(t *testing.T) {
	testPipelineModule := &monagent.PipelineModule{}
	err := json.Unmarshal([]byte(testJSONModule), testPipelineModule)
	if err != nil {
		t.Errorf("test create json module failed %s", err.Error())
		return
	}
	testPipelineInstances, _ := CreatePipelines(testPipelineModule)

	type args struct {
		pipelineModule *monagent.PipelineModule
	}
	tests := []struct {
		name string
		args args
		want []*Pipeline
	}{
		{name: "test", args: args{testPipelineModule}, want: testPipelineInstances},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := CreatePipelines(tt.args.pipelineModule); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePipelineInstances() = %v, want %v", got, tt.want)
			}
		})
	}
}
