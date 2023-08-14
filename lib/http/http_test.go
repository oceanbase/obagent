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

package http

import (
	"testing"
)

func TestProxy(t *testing.T) {
	t.Skip()
	err := SetSocksProxy("127.0.0.1:1081")
	if err != nil {
		t.Errorf("set proxy err: %v", err)
	}
	resp, err := GetGlobalHttpClient().Get("http://127.0.0.1:8080/")
	if err != nil {
		t.Errorf("request failed, %v", err)
	}
	println(resp.Status)
}
