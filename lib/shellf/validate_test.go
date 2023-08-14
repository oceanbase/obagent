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

package shellf

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCommandParameterType_Validate(t *testing.T) {
	type args struct {
		parameterType CommandParameterType
		value         string
	}
	type want struct {
		valid bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "PACKAGE_NAME: name",
			args: args{
				parameterType: packageNameType,
				value:         "obagent",
			},
			want: want{
				valid: true,
			},
		},
		{
			name: "PACKAGE_NAME: name, version",
			args: args{
				parameterType: packageNameType,
				value:         "obagent-2.4.0",
			},
			want: want{
				valid: true,
			},
		},
		{
			name: "PACKAGE_NAME: name, version, build, os, arch",
			args: args{
				parameterType: packageNameType,
				value:         "obagent-2.4.0-1884049.alios7.x86_64",
			},
			want: want{
				valid: true,
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			valid := tt.args.parameterType.Validate(tt.args.value)
			So(valid, ShouldEqual, tt.want.valid)
		})
	}
}
