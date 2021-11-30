//go:build darwin
// +build darwin

package log_tailer

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
