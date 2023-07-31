package command

import (
	"context"
	"testing"
)

func TestInput(t *testing.T) {
	ctx := context.Background()
	input := NewInput(ctx, "arg")
	if input.Context() != ctx || input.Param() != "arg" || input.Annotation() == nil {
		t.Error("input init wrong")
	}
}
