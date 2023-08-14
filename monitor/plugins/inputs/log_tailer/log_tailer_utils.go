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

package log_tailer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/file"
	"github.com/oceanbase/obagent/lib/log_analyzer"
)

// findFilesAndSortByMTime Gets the files in the path that match the
// RE & file size (sorted from old to new by modification time)
func findFilesAndSortByMTime(
	ctx context.Context,
	dir, fileName string,
	start, end time.Time,
	getFileTime file.GetFileTimeFunc,
) ([]os.FileInfo, error) {
	ctxLog := log.WithContext(ctx).WithFields(log.Fields{
		"dir":      dir,
		"fileName": fileName,
		"start":    start,
		"end":      end,
	})
	mTime := -1 * end.Sub(start)
	matchedFiles, err := libFile.FindFilesByRegexAndMTime(ctx, dir, fileName, matchString, mTime, end, getFileTime)
	if err != nil {
		return nil, err
	}
	sort.Sort(ByMTime(matchedFiles))

	ctxLog.WithField("matchedFiles", matchedFiles).Info("invoke FindFilesByRegexAndMTime")
	return matchedFiles, nil
}

// ByMTime  implements sort.Interface for []os.FileInfo based on ModTime().
type ByMTime []os.FileInfo

func (a ByMTime) Len() int           { return len(a) }
func (a ByMTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByMTime) Less(i, j int) bool { return a[i].ModTime().Before(a[j].ModTime()) }

func matchString(reg string, content string) (matched bool, err error) {
	regMatched, err := filepath.Match(reg, content)
	if err != nil {
		return false, err
	}
	regWithTimestampMatched, err := filepath.Match(fmt.Sprintf("%s.[0-9]*", reg), content)
	if err != nil {
		return false, err
	}

	return regMatched || regWithTimestampMatched, nil
}

type processQueue struct {
	queue    []*logFileInfo
	mutex    sync.Mutex
	capacity int
}

func NewProcessQueue(cap int) *processQueue {
	return &processQueue{
		queue:    make([]*logFileInfo, 0),
		mutex:    sync.Mutex{},
		capacity: cap,
	}
}

func (p *processQueue) getLen() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return len(p.queue)
}

func (p *processQueue) getQueueLen() int {
	return len(p.queue)
}

func (p *processQueue) getHead() *logFileInfo {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.getQueueLen() == 0 {
		return nil
	}
	return p.queue[0]
}

func (p *processQueue) getTail() *logFileInfo {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.getQueueLen() == 0 {
		return nil
	}
	return p.queue[p.getQueueLen()-1]
}

func (p *processQueue) popHead() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.getQueueLen() == 0 {
		return
	}
	p.queue = p.queue[1:]
}

func (p *processQueue) pushBack(info *logFileInfo) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.getQueueLen() >= p.capacity {
		return errors.New("queue is full")
	}

	p.queue = append(p.queue, info)
	return nil
}

func (p *processQueue) setHeadIsRenameTrue() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.getQueueLen() == 0 {
		return
	}

	p.queue[0].isRenamed = true
}

func (p *processQueue) getHeadIsRename() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.getQueueLen() == 0 {
		return false
	}

	return p.queue[0].isRenamed
}

func closeFile(ctx context.Context, fd *os.File) {
	ctxLog := log.WithContext(ctx)
	if fd != nil {
		err := fd.Close()
		if err != nil {
			ctxLog.WithField("fileName", fd.Name()).WithError(err).Warn("failed to close file")
		}
	}
}

func closeFiles(ctx context.Context, fds []*os.File) {
	if fds != nil {
		for _, fd := range fds {
			closeFile(ctx, fd)
		}
	}
}

func checkAndOpenFile(ctx context.Context, logFileRealPath string) (*os.File, error) {
	isFileExists, err := libFile.FileExists(logFileRealPath)
	if err != nil {
		return nil, err
	}
	if !isFileExists {
		return nil, nil
	}
	newFileDesc, err := os.OpenFile(logFileRealPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, errors.Errorf("open logFile %s failed", logFileRealPath)
	}
	return newFileDesc, nil
}

func isSameFile(f1 *os.File, f2 *os.File) (bool, error) {
	f1Stat, err := f1.Stat()
	if err != nil {
		return false, err
	}
	f2Stat, err := f2.Stat()
	if err != nil {
		return false, err
	}

	return os.SameFile(f1Stat, f2Stat), nil
}

type FilterLogFunc func(ctx context.Context, fd *os.File) (isFilter bool)

func getLogsWithinTime(ctx context.Context, conf monagent.TailConfig, start, end time.Time) ([]*os.File, error) {
	ctxLog := log.WithContext(ctx)
	logAnalyzer := log_analyzer.GetLogAnalyzer(conf.LogAnalyzerType, conf.LogSourceType)
	if logAnalyzer == nil {
		return nil, errors.Errorf("get log analyzer failed, logAnalyzerType: %s", conf.LogAnalyzerType)
	}
	matchedFileInfos, err := findFilesAndSortByMTime(ctx, conf.LogDir, conf.LogFileName, start, end, logAnalyzer.GetFileEndTime)
	if err != nil {
		return nil, errors.New("findFilesAndSortByMTime failed")
	}
	matchedFiles := make([]*os.File, 0)
	for _, matchedFileInfo := range matchedFileInfos {
		matchedFileRealPath := fmt.Sprintf("%s/%s", conf.LogDir, matchedFileInfo.Name())
		ctxLog.WithField("matchedFileRealPath", matchedFileRealPath).Info("getLogsWithinTime match file")
		matchedFileDesc, err1 := checkAndOpenFile(ctx, matchedFileRealPath)
		if err1 != nil {
			ctxLog.WithError(err1).Warn("checkAndOpenFile error")
			continue
		}
		matchedFiles = append(matchedFiles, matchedFileDesc)
	}
	return matchedFiles, nil
}

func logFileFilterIn(ctx context.Context, fds []*os.File, fn FilterLogFunc) (matchedFiles, unmatchedFiles []*os.File, err error) {
	for _, fd := range fds {
		if fn(ctx, fd) {
			unmatchedFiles = append(unmatchedFiles, fd)
		} else {
			matchedFiles = append(matchedFiles, fd)
		}
	}

	return matchedFiles, unmatchedFiles, nil
}
