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
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	err := InvalidArgument.NewCode("test", "bad_input").NewError("111").
		WithStack()

	fmt.Println(err.Message())
	err2 := Internal.NewCode("test", "internal").NewError().WithCause(err)
	fmt.Println(err2.Error())

	if HttpCode(err, 200) != 400 {
		t.Error("bad http code")
	}
	if err2.Cause() == nil {
		t.Error("should have cause")
	}
	if err.Stack() == "" {
		t.Error("should have stack")
	}
	if KindByName("UNKNOWN") != Unknown {
		t.Error("KindByName wrong")
	}
}
