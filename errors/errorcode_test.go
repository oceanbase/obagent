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

package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCode_CodeDistinct(t *testing.T) {
	m := map[int]ErrorCode{}
	for _, e := range errorCodes {
		if e2, ok := m[e.Code]; ok {
			assert.False(t, ok,
				"conflict code %v, both used by %v and %v", e.Code, e.key, e2.key)
		}
		m[e.Code] = e
	}
}

func TestErrorCode_KeyDistinct(t *testing.T) {
	m := map[string]ErrorCode{}
	for _, e := range errorCodes {
		if e2, ok := m[e.key]; ok {
			assert.False(t, ok,
				"conflict code %v, both used by %v and %v", e.key, e.Code, e2.Code)
		}
		m[e.key] = e
	}
}
