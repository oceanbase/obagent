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

package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/oceanbase/obagent/lib/file"
	"github.com/shirou/gopsutil/v3/disk"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/errors"
)

var libFile file.File = file.FileImpl{}

func GetRealPath(ctx context.Context, dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", errors.Occur(errors.ErrIllegalArgument)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", errors.Occur(errors.ErrUnexpected, err)
	}
	realPath, err := filepath.EvalSymlinks(absDir)
	if err != nil {
		return "", errors.Occur(errors.ErrUnexpected, err)
	}
	return realPath, nil
}

func GetDiskUsage(ctx context.Context, dir string) (float64, error) {
	ctxLog := log.WithContext(ctx)

	realPath, err := GetRealPath(ctx, dir)
	if err != nil {
		return 0, errors.Occur(errors.ErrUnexpected, err)
	}

	usageStat, err := disk.Usage(realPath)
	if err != nil {
		ctxLog.WithField("realPath", realPath).WithError(err).Error("disk.Usage failed")
		return 0, errors.Occur(errors.ErrUnexpected, err)
	}

	ctxLog.WithFields(log.Fields{
		"usageStat": usageStat,
		"realPath":  realPath,
	}).Info("get usage")

	return usageStat.UsedPercent, nil
}

func FindFilesAndSortByMTime(ctx context.Context, dir, fileRegex string) ([]fileInfoWithPath, error) {
	realPath, err := GetRealPath(ctx, dir)
	if err != nil {
		return nil, err
	}
	matchedFiles, err := FindFilesByRegexAndMTime(ctx, realPath, fileRegex, 0)
	if err != nil {
		return nil, err
	}
	sort.Sort(ByMTime(matchedFiles))

	return matchedFiles, nil
}

func FindFilesByRegexAndMTime(ctx context.Context, dir, fileRegex string, mTime time.Duration) ([]fileInfoWithPath, error) {
	matchedFiles, err := libFile.FindFilesByRegexAndMTime(ctx, dir, fileRegex, matchRegex, mTime, time.Now(), file.GetFileModTime)
	if err != nil {
		return nil, err
	}
	fileInfoWithPathArr := make([]fileInfoWithPath, len(matchedFiles))
	for i, matchedFile := range matchedFiles {
		fileInfoWithPathArr[i] = fileInfoWithPath{
			Path: filepath.Join(dir, matchedFile.Name()),
			Info: matchedFile,
		}
	}
	return fileInfoWithPathArr, nil
}

type fileInfoWithPath struct {
	Path string
	Info os.FileInfo
}

func (f fileInfoWithPath) String() string {
	return f.Info.Name()
}

// ByMTime  implements sort.Interface for []fileInfoWithPath based on ModTime().
type ByMTime []fileInfoWithPath

func (a ByMTime) Len() int           { return len(a) }
func (a ByMTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByMTime) Less(i, j int) bool { return a[i].Info.ModTime().Before(a[j].Info.ModTime()) }

func matchRegex(reg string, content string) (matched bool, err error) {
	matched, err = regexp.MatchString(reg, content)
	return
}
