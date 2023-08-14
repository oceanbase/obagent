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
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/lib/shell"
)

const (
	x86 = "x86_64"
	arm = "aarch64"
)

const (
	redhat = "rhel"
	debian = "debian"
	suse   = "suse"
)

func TestCommandGroup_SelectCommandTemplate(t *testing.T) {
	configString := `commandGroups:
  - name: package.info
    program: "bash"
    timeout: 3m
    commands:
    - case:
        os: rhel
      cmd: "rpm -qi ${PACKAGE_NAME}"
    - case:
        os: debian
      cmd: "dpkg -l ${PACKAGE_NAME}"
    - case:
        os: suse
      cmd: "rpm -qi ${PACKAGE_NAME}"
      program: "bash"
    - default:
      cmd: "rpm -qi ${PACKAGE_NAME}"
    params:
    - name: "PACKAGE_NAME"
      validate: PACKAGE_NAME
`
	config, err := decodeConfig(configString)
	if err != nil {
		t.Error(err)
	}
	group, err := buildCommandGroup(config.CommandGroups[0])
	if err != nil {
		t.Error(err)
	}

	type args struct {
		os   string
		arch string
	}
	type want struct {
		template string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "rhel",
			args: args{
				os:   redhat,
				arch: x86,
			},
			want: want{
				template: "rpm -qi ${PACKAGE_NAME}",
			},
		},
		{
			name: "debian",
			args: args{
				os:   debian,
				arch: x86,
			},
			want: want{
				template: "dpkg -l ${PACKAGE_NAME}",
			},
		},
		{
			name: "suse",
			args: args{
				os:   suse,
				arch: arm,
			},
			want: want{
				template: "rpm -qi ${PACKAGE_NAME}",
			},
		},
		{
			name: "default",
			args: args{
				os:   "unknown",
				arch: arm,
			},
			want: want{
				template: "rpm -qi ${PACKAGE_NAME}",
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			template := group.SelectCommandTemplate(tt.args.os, tt.args.arch)
			assert.Equal(t, tt.want.template, template.Template)
		})
	}
}

func TestCommandTemplate_Instantiate(t *testing.T) {
	type args struct {
		configString string
	}
	type want struct {
		program shell.Program
		user    string
		timeout time.Duration
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "use group config",
			args: args{
				configString: `commandGroups:
  - name: package.info
    program: sh
    user: admin
    timeout: 3m
    commands:
    - default:
      cmd: "rpm -qi oceanbase"
`,
			},
			want: want{
				program: shell.Sh,
				user:    "admin",
				timeout: 3 * time.Minute,
			},
		},
		{
			name: "use branch config",
			args: args{
				configString: `commandGroups:
  - name: package.info
    program: sh
    user: admin
    timeout: 3m
    commands:
    - default:
      cmd: "rpm -qi oceanbase"
      program: bash
      user: root
`,
			},
			want: want{
				program: shell.Bash,
				user:    "root",
				timeout: 3 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			config, err := decodeConfig(tt.args.configString)
			if err != nil {
				t.Error(err)
			}
			group, err := buildCommandGroup(config.CommandGroups[0])
			So(err, ShouldBeNil)
			template := group.SelectCommandTemplate(redhat, x86)
			command, err := template.Instantiate(nil)
			So(err, ShouldBeNil)
			So(command.User(), ShouldEqual, tt.want.user)
			So(command.Program(), ShouldEqual, tt.want.program)
			So(command.Timeout(), ShouldEqual, tt.want.timeout)
		})
	}
}

func TestShellShelf_GetCommand(t *testing.T) {
	config, err := decodeConfig(configString)
	if err != nil {
		t.Error(err)
	}
	groupMap, err := buildCommandGroupMap(config.CommandGroups)
	if err != nil {
		t.Error(err)
	}
	shelf.groups = &groupMap

	command, err := Shelf().GetCommand("package.info", debian, x86, map[string]string{
		"PACKAGE_NAME": "oceanbase",
	})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "dpkg -l oceanbase", command.Cmd())
}

func decodeConfig(configString string) (*ShellfConfig, error) {
	r := strings.NewReader(configString)
	config := new(ShellfConfig)
	err := yaml.NewDecoder(r).Decode(config)
	return config, err
}

var configString = `commandGroups:
  - name: package.info
    program: "bash"
    timeout: 3m
    commands:
    - case: # 这里的 case 和下面的 default 语义类似 switch-case 语句，会对不同的 os 使用不同的命令
        os: rhel # RedHat 系（RedHat, CentOS, AliOS）
      cmd: "rpm -qi ${PACKAGE_NAME}"
    - case:
        os: debian # Debian 系（Debian, Ubuntu）
      cmd: "dpkg -l ${PACKAGE_NAME}"
    - case:
        os: suse # SUSE 操作系统
      cmd: "rpm -qi ${PACKAGE_NAME}"
      program: "bash" # 此处的 program 定义可以覆盖外层的 program
    - default: # 没有匹配以上规则的会走到 default
      cmd: "rpm -qi ${PACKAGE_NAME}"
    params:
    - name: "PACKAGE_NAME" # 参数名与命令模板中的占位符相对应
      validate: PACKAGE_NAME # 该参数期望的格式是标识符（字母、数字、下划线），会根据这个格式对 PACKAGE_NAME 参数进行校验

  - name: package.install
    program: "sh"
    commands:
    - case: # 这里会同时匹配 os 和 arch 两个规则，两者是「且」的关系
        os: debian
        arch: x86_64
      cmd: "alien -k -i ${PACKAGE_FILE}"
    - case:
        os: debian
        arch: aarch64
      cmd: "alien -k -i ${PACKAGE_FILE} --target=arm64"
    - case: # 这里只写了 os 一个 key，那么另一个 key 是 "any"，即匹配 os=rhel, arch=* 的情况
        os: rhel
      cmd: "rpm -Uvh ${PACKAGE_FILE}"
    - default:
      cmd: "rpm -Uvh ${PACKAGE_FILE}"
    params:
    - name: "PACKAGE_FILE"
      validate: PATH # 该参数期望的格式是文件路径，会根据这个格式对 PACKAGE_FILE 参数进行校验
`
