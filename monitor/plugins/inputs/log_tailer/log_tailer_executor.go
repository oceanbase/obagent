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

package log_tailer

import (
	"bufio"
	"context"
	"github.com/oceanbase/obagent/monitor/plugins/common"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/file"
	"github.com/oceanbase/obagent/lib/log_analyzer"
	"github.com/oceanbase/obagent/monitor/message"
)

var libFile file.File = file.FileImpl{}

// LogTailerExecutor Log tracing for each tailConf
type LogTailerExecutor struct {
	tailConf         monagent.TailConfig
	recoveryConf     monagent.RecoveryConfig
	out              chan<- []*message.Message
	fileProcessQueue *processQueue
	stopFlagMutex    *sync.Mutex
	toBeStopped      chan bool
	isStoppedFlag    bool
	tailLineCount    uint64
}

func NewLogTailerExecutor(
	tailConf monagent.TailConfig,
	recoveryConf monagent.RecoveryConfig,
	toBeStopped chan bool,
	out chan<- []*message.Message,
) *LogTailerExecutor {
	return &LogTailerExecutor{
		tailConf:         tailConf,
		recoveryConf:     recoveryConf,
		out:              out,
		fileProcessQueue: NewProcessQueue(10),
		stopFlagMutex:    &sync.Mutex{},
		toBeStopped:      toBeStopped,
	}
}
func (l *LogTailerExecutor) isStopped() bool {
	l.stopFlagMutex.Lock()
	defer l.stopFlagMutex.Unlock()
	return l.isStoppedFlag
}

func (l *LogTailerExecutor) markAsStopped(ctx context.Context) {
	l.stopFlagMutex.Lock()
	l.isStoppedFlag = true
	l.stopFlagMutex.Unlock()

	queueHead := l.fileProcessQueue.getHead()
	ctxLog := log.WithContext(ctx)
	err := storeLastPosition(ctx, l.recoveryConf, queueHead)
	if err != nil {
		ctxLog.WithError(err).Warnf("failed to store last position，logFileInfo:%+v", queueHead)
	}
}

func (l *LogTailerExecutor) checkAndStoreLastPosition(ctx context.Context) {
	if !l.recoveryConf.Enabled || l.recoveryConf.TriggerStoreThreshold == 0 {
		return
	}

	ctxLog := log.WithContext(ctx)
	l.tailLineCount++
	queueHead := l.fileProcessQueue.getHead()
	if l.tailLineCount >= l.recoveryConf.TriggerStoreThreshold {
		err := storeLastPosition(ctx, l.recoveryConf, queueHead)
		if err != nil {
			ctxLog.WithError(err).Warn("failed to store last position")
		}
		l.tailLineCount = 0
	}
}

func (l *LogTailerExecutor) processLogByLine(
	ctx context.Context,
	fileInfo *logFileInfo,
) error {
	fd := fileInfo.fileDesc
	ctxLog := log.WithContext(ctx).WithField("fileName", fd.Name())
	analyzer := log_analyzer.GetLogAnalyzer(fileInfo.logAnalyzerType, fileInfo.fileName)
	fdScanner := bufio.NewScanner(fd)
	fdScanner.Split(file.ScanLines)
	lineHandler := func() (string, bool, error) {
		select {
		case _, isOpen := <-l.toBeStopped:
			if !isOpen {
				l.markAsStopped(ctx)
				ctxLog.Info("stop process line")
				return "", false, nil
			}
		default:
		}
		ok := fdScanner.Scan()
		if !ok {
			err := fdScanner.Err()
			if err != nil {
				ctxLog.WithError(err).Warn("failed to scan")
			}
			return "", false, err
		}
		text := fdScanner.Text()
		fileInfo.fileOffset += int64(len(text))
		return text, true, nil
	}
	msgConsumer := func(m *message.Message) bool {
		select {
		case _, isOpen := <-l.toBeStopped:
			if !isOpen {
				l.markAsStopped(ctx)
				ctxLog.Info("stop consume line")
				return false
			}
		default:
		}
		m.AddTag(common.LogSourceType, l.tailConf.LogSourceType)
		m.AddTag(common.AbsLogFileName, filepath.Join(l.tailConf.LogDir, l.tailConf.LogFileName))
		fileInfo.offsetLineLogAt = m.GetTime()
		l.out <- []*message.Message{m}
		l.checkAndStoreLastPosition(ctx)
		return true
	}
	return log_analyzer.ParseLines(analyzer, lineHandler, msgConsumer)
}

func (l *LogTailerExecutor) handleFileQueue(ctx context.Context) error {
	logFileRealPath := l.tailConf.GetLogFileRealPath()
	ctxLog := log.WithContext(ctx).WithFields(log.Fields{"tailConf": l.tailConf, "logFileRealPath": logFileRealPath})
	if logFileRealPath == "" {
		ctxLog.Warn("invalid logFileRealPath")
		return nil
	}

	var queueHead *logFileInfo
	for {
		queueHead = l.fileProcessQueue.getHead()
		select {
		case _, isOpen := <-l.toBeStopped:
			if !isOpen {
				l.markAsStopped(ctx)
				ctxLog.Info("stop handleFileProcessQueue")
				return nil
			}
		default:
		}
		if queueHead != nil {
			// Each time the file at the head of the queue is read,
			// if the old file is not read, then the fd of the old file is always obtained
			err := l.processLogByLine(ctx, queueHead)
			if err != nil {
				ctxLog.WithError(err).Warn("processLogByLine error")
			}
			// The file has been renamed (written out) and the loop has read the entire file, then close the file & dequeue
			if l.fileProcessQueue.getLen() > 1 || l.fileProcessQueue.getHeadIsRename() {
				fileName := queueHead.fileDesc.Name()
				ctxLog.WithField("fileName", fileName).Info("pop head from fileProcessQueue")
				// pop and then close to prevent other goroutine from getting the closed file
				l.fileProcessQueue.popHead()
				closeFile(ctx, queueHead.fileDesc)
			}
		} else {
			time.Sleep(l.tailConf.ProcessLogInterval)
		}

		time.Sleep(l.tailConf.ProcessLogInterval)
	}
}

func (l *LogTailerExecutor) WatchFile(ctx context.Context) error {
	logFileRealPath := l.tailConf.GetLogFileRealPath()
	ctxLog := log.WithContext(ctx).WithField("fileRealPath", logFileRealPath)
	if logFileRealPath == "" {
		ctxLog.Warn("invalid logFileRealPath")
		return nil
	}
	ctxLog.Infof("start watching")

	var (
		latestFileDesc *os.File
		err            error
	)
	for {
		closeFile(ctx, latestFileDesc)
		select {
		case _, isOpen := <-l.toBeStopped:
			if !isOpen {
				l.markAsStopped(ctx)
				ctxLog.Info("stop WatchFile")
				return nil
			}
		default:
			time.Sleep(l.tailConf.ProcessLogInterval)
		}

		latestFileDesc, err = checkAndOpenFile(ctx, logFileRealPath)
		if err != nil {
			ctxLog.WithError(err).Warn("failed to open latest log file")
			continue
		}
		if latestFileDesc == nil {
			continue
		}
		queueTail := l.fileProcessQueue.getTail()
		if queueTail == nil {
			if latestFileDesc != nil {
				err1 := l.initQueue(ctx)
				if err1 != nil {
					ctxLog.WithError(err1).Warn("initQueue failed")
				}
			}
			continue
		}
		isSameLog, err1 := isSameFile(queueTail.fileDesc, latestFileDesc)
		if err1 != nil {
			ctxLog.WithError(err1).Warnf("failed to compare two file fd1:%s, fd2:%s", queueTail.fileDesc.Name(), latestFileDesc.Name())
			continue
		}
		if isSameLog {
			continue
		}
		err = l.handleWatchedNewLogs(ctx, queueTail.fileDesc)
		if err != nil {
			ctxLog.WithError(err).Warn("handleWatchedNewLogs failed")
		}
	}
}

func (l *LogTailerExecutor) initQueue(ctx context.Context) error {
	logFileRealPath := l.tailConf.GetLogFileRealPath()
	ctxLog := log.WithContext(ctx).WithField("fileRealPath", logFileRealPath)
	logFileDesc, err := checkAndOpenFile(ctx, logFileRealPath)
	if err != nil {
		closeFile(ctx, logFileDesc)
		return err
	}
	ctxLog.Info("add logFileInfo to empty queue")
	return l.fileProcessQueue.pushBack(&logFileInfo{
		logSourceType:   l.tailConf.LogSourceType,
		logAnalyzerType: l.tailConf.LogAnalyzerType,
		fileName:        l.tailConf.LogFileName,
		fileDesc:        logFileDesc,
		isRenamed:       false,
	})
}

func (l *LogTailerExecutor) handleWatchedNewLogs(ctx context.Context, queueTailFd *os.File) error {
	ctxLog := log.WithContext(ctx)
	queueTailFileStat, err := libFile.GetFileStatInfo(ctx, queueTailFd)
	if err != nil {
		return errors.New("queueTail GetFileStatInfo failed")
	}
	queueTailCTime := queueTailFileStat.CreateTime()
	start, end := time.Date(queueTailCTime.Year(), queueTailCTime.Month(), queueTailCTime.Day(),
		queueTailCTime.Hour(), queueTailCTime.Minute(), queueTailCTime.Second(), 0, time.Local), time.Now()
	ctxLog.WithFields(log.Fields{
		"start":                 start,
		"end":                   end,
		"queueTailFileStat.ino": queueTailFileStat.FileId(),
	}).Info("get file change events")

	newLogs, err := l.getWatchedNewLogs(ctx, queueTailFd, start, end)
	if err != nil {
		return errors.New("getWatchedNewLogs failed")
	}
	newLogsLength := len(newLogs)
	for i, newLog := range newLogs {
		if newLog == nil {
			continue
		}
		matchedLogFileInfo := &logFileInfo{
			logSourceType:   l.tailConf.LogSourceType,
			logAnalyzerType: l.tailConf.LogAnalyzerType,
			fileName:        l.tailConf.LogFileName,
			fileDesc:        newLog,
			isRenamed:       false,
		}
		if i != newLogsLength-1 {
			matchedLogFileInfo.isRenamed = true
		}
		err1 := l.fileProcessQueue.pushBack(matchedLogFileInfo)
		if err1 != nil {
			ctxLog.WithField("fileName", l.tailConf.LogFileName).Warn("handleWatchedNewLogs enqueue failed")
		}
		l.fileProcessQueue.setHeadIsRenameTrue()
	}
	return nil
}

// getWatchedNewLogs Obtain the log files added within the time range
func (l *LogTailerExecutor) getWatchedNewLogs(ctx context.Context, queueTailFd *os.File, start, end time.Time) ([]*os.File, error) {
	ctxLog := log.WithContext(ctx).WithField("tailConf", l.tailConf)
	newLogFiles, err := getLogsWithinTime(ctx, l.tailConf, start, end)
	if err != nil {
		closeFiles(ctx, newLogFiles)
		return nil, errors.New("getLogsWithinTime failed")
	}
	filterFunc := func(ctx context.Context, fd *os.File) (isFilter bool) {
		ctxLog1 := log.WithContext(ctx)
		isSame, err3 := isSameFile(queueTailFd, fd)
		if err3 != nil {
			ctxLog1.WithField("fd", fd).Warn("check isSameFile failed")
			return true
		}

		return isSame
	}
	matchedFiles, unmatchedFiles, err := logFileFilterIn(ctx, newLogFiles, filterFunc)
	defer closeFiles(ctx, unmatchedFiles)
	if err != nil {
		ctxLog.WithFields(log.Fields{"start": start, "end": end}).Warn("logFileFilterIn failed")
		closeFiles(ctx, matchedFiles)
		return nil, err
	}

	ctxLog.WithField("matchedFiles", matchedFiles).Debug("after logFileFilterIn")
	return matchedFiles, nil
}

// TailLog starts to tail given log
// If you are opening it for the first time, read the latest log_file and then monitor the file changes
func (l *LogTailerExecutor) TailLog(ctx context.Context) error {
	ctxLog := log.WithContext(ctx).WithField("tailConf", l.tailConf)

	if l.recoveryConf.Enabled {
		lastLogFileInfo, err1 := loadLastPosition(ctx, l.recoveryConf, l.tailConf)
		if err1 != nil {
			ctxLog.WithError(err1).Warn("loadLastPosition failed")
		}
		if lastLogFileInfo != nil {
			l.fileProcessQueue.queue = []*logFileInfo{lastLogFileInfo}
		}
	}

	ctxLog.WithField("initLogQueue", l.fileProcessQueue.queue).Info("build fileProcessQueueMap with initLogQueue")

	go func() {
		err1 := l.handleFileQueue(ctx)
		if err1 != nil {
			ctxLog.WithError(err1).Warn("failed to handleFileProcessQueue")
			return
		}
	}()

	go func() {
		err1 := l.WatchFile(ctx)
		if err1 != nil {
			ctxLog.WithField("logFileRealPath", l.tailConf.LogFileName).WithError(err1).Warn("failed to watch file")
			return
		}
	}()

	return nil
}
