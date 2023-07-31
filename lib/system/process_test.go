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
