package trace

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	agentlog "github.com/oceanbase/obagent/log"
)

func TestGetTraceId_WithTraceId(t *testing.T) {
	req := &http.Request{
		Header: http.Header{},
	}
	traceId := "abcdefg"
	req.Header.Add(TraceIdHeader, traceId)
	if got := GetTraceId(req); got != traceId {
		t.Errorf("GetTraceId() = %v, want %v", got, traceId)
	}

	ctx := ContextWithTraceId(req)
	val := ctx.Value(agentlog.TraceIdKey{})
	assert.NotNil(t, val)
	ctxTraceId, _ := val.(string)
	assert.Equal(t, traceId, ctxTraceId)
}

func TestGetTraceId_GetRandomTraceId(t *testing.T) {
	req := &http.Request{
		Header: http.Header{},
	}
	if got := GetTraceId(req); got == "" {
		t.Errorf("GetTraceId() = %v, want not nil", got)
	}
}
