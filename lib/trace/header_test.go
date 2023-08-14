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
