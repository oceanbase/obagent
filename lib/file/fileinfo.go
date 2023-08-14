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

package file

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

func GetFileInfo(f *os.File) (*FileInfoEx, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return toFileInfoEx(info), nil
}
