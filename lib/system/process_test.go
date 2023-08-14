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

package system

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseNetstatLine(t *testing.T) {
	type args struct {
		netstatLine string
	}
	type want struct {
		port int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "tcp4",
			args: args{
				netstatLine: "tcp        0      0 0.0.0.0:62888           0.0.0.0:*               LISTEN      104668/pos_proxy",
			},
			want: want{
				port: 62888,
			},
		},
		{
			name: "tcp4",
			args: args{
				netstatLine: "tcp6       0      0 :::62888                :::*                    LISTEN      104668/pos_proxy",
			},
			want: want{
				port: 62888,
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			port, ok := parseNetstatLine(tt.args.netstatLine)
			So(ok, ShouldBeTrue)
			So(port, ShouldEqual, tt.want.port)
		})
	}
}
