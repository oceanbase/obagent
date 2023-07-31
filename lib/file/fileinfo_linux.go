//go:build linux
// +build linux

package file

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
		devId:      sysInfo.Dev,
		createTime: time.Unix(sysInfo.Ctim.Unix()),
	}
}
