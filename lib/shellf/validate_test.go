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
