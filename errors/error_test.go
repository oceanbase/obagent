// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package errors

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorOccur(t *testing.T) {
	e := ErrUnexpected
	args := "unexpected"
	err := Occur(e, args)
	message := err.Error()
	assert.Contains(t, message, strconv.Itoa(e.Code))
	assert.Contains(t, message, args)
}
