package log

import (
        "os"
        "time"
)

type FileInfoEx struct {
        os.FileInfo
        fileId     uint64
        devId      uint64
        createTime time.Time
}

func (f *FileInfoEx) FileId() uint64 {
        return f.fileId
}

func (f *FileInfoEx) DevId() uint64 {
        return f.devId
}

func (f *FileInfoEx) CreateTime() time.Time {
        return f.createTime
}

func FileInfo(f *os.File) (*FileInfoEx, error) {
        info, err := f.Stat()
        if err != nil {
                return nil, err
        }
        return toFileInfoEx(info), nil
}
