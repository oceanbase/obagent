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

//go:build darwin
// +build darwin

package log

import (
	"os"
	"syscall"
	"time"
)

func toFileInfoEx(info os.FileInfo) *FileInfoEx {
	sysInfo, _ := info.Sys().(*syscall.Stat_t)
	return &FileInfoEx{
		FileInfo:   info,
		fileId:     sysInfo.Ino,
		devId:      uint64(sysInfo.Dev),
		createTime: time.Unix(sysInfo.Ctimespec.Unix()),
	}
}
