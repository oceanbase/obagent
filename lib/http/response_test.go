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
