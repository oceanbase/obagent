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

package mask

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskCommandPassword(t *testing.T) {
	type args struct {
		before string
		after  string
	}
	tests := []struct {
		args args
	}{
		{
			args: args{
				before: "python -uroot@sys -l some word",
				after:  "python -uroot@sys -l some word",
			},
		},
		{
			args: args{
				before: "python -uroot@sys --password='somepassword' -l some word",
				after:  "python -uroot@sys --password=xxx -l some word",
			},
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.args.after, maskCommandPassword(tt.args.before))
	}
}

func TestMaskScriptPassword(t *testing.T) {
	type args struct {
		before string
		after  string
	}
	tests := []struct {
		args args
	}{
		{
			args: args{
				before: "python ./bin/import_time_zone_info.py -h10.10.10.10 -P2881 -p123456 -tmysql1 -f ./etc/timezone_V1.log",
				after:  "python ./bin/import_time_zone_info.py -h10.10.10.10 -P2881 -pxxx -tmysql1 -f ./etc/timezone_V1.log",
			},
		},
		{
			args: args{
				before: "python ./bin/import_time_zone_info.py -h10.10.10.10 -P2881 -p=123456 -tmysql1 -f ./etc/timezone_V1.log",
				after:  "python ./bin/import_time_zone_info.py -h10.10.10.10 -P2881 -p=xxx -tmysql1 -f ./etc/timezone_V1.log",
			},
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.args.after, maskScriptPassword(tt.args.before))
	}
}

func TestMaskMysqlPassword(t *testing.T) {
	before := "mysql -h127.1 -P2881 -uocp_monitor -p'Hello123' -e \"select 1 from dual;\""
	after := "mysql -h127.1 -P2881 -uocp_monitor -pxxx -e \"select 1 from dual;\""
	assert.Equal(t, after, maskMysqlPassword(before))
}

func TestMaskMysqlDSN(t *testing.T) {
	type args struct {
		before string
		after  string
	}
	tests := []struct {
		args args
	}{
		{
			args: args{
				before: "root:debug@tcp(127.0.0.1:2881)/oceanbase?interpolateParams=true",
				after:  "root:xxx@tcp(127.0.0.1:2881)/oceanbase?interpolateParams=true",
			},
		},
		{
			args: args{
				before: "root@sys:debug@tcp(127.0.0.1:2881)/oceanbase?interpolateParams=true",
				after:  "root@sys:xxx@tcp(127.0.0.1:2881)/oceanbase?interpolateParams=true",
			},
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.args.after, maskMysqlDSN(tt.args.before))
	}
}

func TestMaskDumpBackup(t *testing.T) {
	before := "./ob_admin dump_backup -d 'oss:/xxx' -s 'host=xxx&access_id=123&access_key=xyz'"
	after := "./ob_admin dump_backup -d 'oss:/xxx' -s 'host=xxx&access_id=xxx&access_key=xxx'"
	assert.Equal(t, after, maskDumpBackup(before))
}
