package http

import (
	"encoding/json"
	"fmt"
	"testing"
)

type S struct {
	A string
	B int
}

func TestJson(t *testing.T) {
	resp := OcpAgentResponse{
		Data: S{},
	}
	err := json.Unmarshal([]byte(`{"Data":{"A":"11","B":123}}`), &resp)
	if err != nil {
		t.Error(err)
	}
	if resp.Data.(S).A != "11" || resp.Data.(S).B != 123 {
		t.Error("data wrong")
	}

	ps := &S{}
	resp = OcpAgentResponse{
		Data: ps,
	}
	err = json.Unmarshal([]byte(`{"Data":{"A":"22","B":321}}`), &resp)
	if err != nil {
		t.Error(err)
	}
	if resp.Data.(*S).A != "22" || ps.B != 321 {
		t.Error("data wrong")
	}

	resp = OcpAgentResponse{}
	err = json.Unmarshal([]byte(`{"Data":{"A":"33","B":999}}`), &resp)
	if err != nil {
		t.Error(err)
	}
	if _, ok := resp.Data.(map[string]interface{}); !ok {
		t.Error("should be a map")
	}
	fmt.Println(resp.Data)
}
