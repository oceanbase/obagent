package mgragent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/lib/command"
	http2 "github.com/oceanbase/obagent/lib/http"
	path2 "github.com/oceanbase/obagent/lib/path"
)

type S struct {
	agent.TaskToken
	A string
}

func TestAsyncCommandHandler(t *testing.T) {
	os.MkdirAll(path2.TaskStoreDir(), 0755)
	defer os.RemoveAll(path2.TaskStoreDir())

	h := asyncCommandHandler(command.WrapFunc(func(s S) S {
		s.A = s.A + s.A
		return s
	}))
	req, _ := http.NewRequest("POST", "/xxx", strings.NewReader(`{"A":"a", "taskToken":"token12345"}`))
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Keys = map[string]interface{}{common.TraceIdKey: "a"}
	h(ctx)
	resp := ctx.Keys[common.OcpAgentResponseKey].(http2.OcpAgentResponse)
	if !resp.Successful || resp.Status != 200 {
		t.Errorf("Fail %+v", resp)
		return
	}
	tokenResult := resp.Data.(TaskTokenResult)
	if tokenResult.TaskToken != "token12345" {
		t.Errorf("bad result %+v", tokenResult)
		return
	}
	result, ok := taskExecutor.WaitResult(command.ExecutionTokenFromString("token12345"))
	if !ok {
		t.Error("wait result failed")
		return
	}
	s := result.Result.(S)
	if s.A != "aa" {
		t.Errorf("bad result %+v", s)
	}
}

func TestAsyncCommandHandler2(t *testing.T) {
	os.MkdirAll(path2.TaskStoreDir(), 0755)
	defer os.RemoveAll(path2.TaskStoreDir())

	h := asyncCommandHandler(command.WrapFunc(func(ctx context.Context, s S) (S, error) {
		s.A = s.A + s.A
		return s, nil
	}))
	req, _ := http.NewRequest("POST", "/xxx", strings.NewReader(`{"A":"a", "taskToken":"token12345"}`))
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Keys = map[string]interface{}{common.TraceIdKey: "a"}
	h(ctx)
	resp := ctx.Keys[common.OcpAgentResponseKey].(http2.OcpAgentResponse)
	if !resp.Successful || resp.Status != 200 {
		t.Errorf("Fail %+v", resp)
		return
	}
	tokenResult := resp.Data.(TaskTokenResult)
	if tokenResult.TaskToken != "token12345" {
		t.Errorf("bad result %+v", tokenResult)
		return
	}
	result, ok := taskExecutor.WaitResult(command.ExecutionTokenFromString("token12345"))
	if !ok {
		t.Error("wait result failed")
		return
	}
	s := result.Result.(S)
	if s.A != "aa" {
		t.Errorf("bad result %+v", s)
	}
}
