package log_query

import (
	"bufio"
	"context"
	"github.com/oceanbase/obagent/monitor/plugins/common"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/mgragent"
	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/file"
	"github.com/oceanbase/obagent/lib/log_analyzer"
	"github.com/oceanbase/obagent/monitor/message"
)

var libFile file.File = file.FileImpl{}

var GlobalLogQuerier *LogQuerier

const maxScanBufferSize = 128 * 1024

const minPosGap = 512 * 1024

type LogQuerier struct {
	conf      *mgragent.LogQueryConfig
	mutex     sync.Mutex
	minPosGap int64
}

func NewLogQuerier(conf *mgragent.LogQueryConfig) *LogQuerier {
	return &LogQuerier{
		conf:      conf,
		mutex:     sync.Mutex{},
		minPosGap: minPosGap,
	}
}

func InitLogQuerierConf(conf *mgragent.LogQueryConfig) error {
	if GlobalLogQuerier == nil {
		GlobalLogQuerier = NewLogQuerier(conf)
	}
	return nil
}

func UpdateLogQuerierConf(conf *mgragent.LogQueryConfig) error {
	if GlobalLogQuerier == nil {
		return errors.New("GlobalLogQuerier is nil")
	}
	return GlobalLogQuerier.UpdateConf(conf)
}

func (l *LogQuerier) UpdateConf(conf *mgragent.LogQueryConfig) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.conf = conf
	return nil
}

func (l *LogQuerier) GetConf() mgragent.LogQueryConfig {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return *l.conf
}

func (l *LogQuerier) Query(ctx context.Context, logQuery *LogQuery) (*Position, error) {
	defer close(logQuery.logEntryChan)
	ctxLog := log.WithContext(ctx).WithField("params", *logQuery.queryLogParams)

	matchedFiles, err := l.getMatchedFiles(ctx, logQuery)
	defer func() {
		for _, matchedFile := range matchedFiles {
			matchedFile.FileDesc.Close()
		}
	}()
	if err != nil {
		ctxLog.WithError(err).Error("getMatchedFiles failed")
		return nil, err
	}
	ctxLog.WithField("matchedFiles:", matchedFiles).Info("queryLog get matchedFiles")
	var lastPos *Position

	for _, matchedFile := range matchedFiles {
		select {
		case <-ctx.Done():
			ctxLog.Info("timeout exceed")
			return lastPos, nil
		default:
		}
		fileInfo := &FileInfo{
			FileName:   matchedFile.FileInfo.Name(),
			FileId:     matchedFile.FileId,
			FileOffset: matchedFile.FileOffset,
		}
		lastPos, err = l.queryLogByLine(ctx, fileInfo, matchedFile.FileDesc, logQuery, matchedFile.LogAnalyzer)
		if err != nil {
			ctxLog.WithError(err).Error("queryLogByLine failed")
			return lastPos, err
		}
		if logQuery.IsExceedLimit() {
			return lastPos, nil
		}
	}
	return lastPos, nil
}

func (l *LogQuerier) queryLogByLine(
	ctx context.Context,
	fileInfo *FileInfo,
	reader io.Reader,
	logQuery *LogQuery,
	logAnalyzer log_analyzer.LogAnalyzer,
) (*Position, error) {
	ctxLog := log.WithContext(ctx).WithFields(log.Fields{
		"fileName": fileInfo.FileName,
	})
	fileOffset := fileInfo.FileOffset
	fdScanner := bufio.NewScanner(reader)
	fdScanner.Split(file.ScanLines)
	fdScanner.Buffer(make([]byte, bufio.MaxScanTokenSize), maxScanBufferSize)
	errCount := 0

	prevLogEntry := LogEntry{
		FileId:     fileInfo.FileId,
		FileOffset: fileOffset,
	}
	for fdScanner.Scan() {
		select {
		case <-ctx.Done():
			ctxLog.Info("timeout exceed")
			return &Position{
				FileId:     prevLogEntry.FileId,
				FileOffset: prevLogEntry.FileOffset,
			}, nil
		default:
		}
		lineBytes := fdScanner.Bytes()
		fileOffset += int64(len(lineBytes))
		if len(lineBytes) == 0 {
			continue
		}
		if errCount > l.GetConf().ErrCountLimit {
			ctxLog.Info("exceed error count limit")
			break
		}
		logLineInfo, isNewLine := logAnalyzer.ParseLine(string(lineBytes))

		isMatchedByLogAt, isSkip, err1 := l.isMatchByLogAt(ctx, logLineInfo, logQuery)
		if err1 != nil {
			errCount++
			ctxLog.WithError(err1).Error("check isMatchByLogAt failed")
			continue
		}

		if isSkip {
			break
		}

		isMatchedByKeyword, err1 := l.isMatchByKeyword(ctx, lineBytes, logQuery)
		if err1 != nil {
			errCount++
			ctxLog.WithError(err1).Error("check isMatchByKeyword failed")
			continue
		}

		curLineBytes := make([]byte, len(lineBytes))
		copy(curLineBytes, lineBytes)
		if isNewLine {
			if prevLogEntry.isMatched {
				logQuery.SendLogEntry(prevLogEntry)
				if logQuery.IsExceedLimit() {
					return &Position{
						FileId:     prevLogEntry.FileId,
						FileOffset: prevLogEntry.FileOffset,
					}, nil
				}
			}

			logLevel, _ := logLineInfo.GetTag(common.Level)
			isMatchedByLogLevel, err1 := l.isMatchByLogLevel(ctx, logLevel, logQuery)
			if err1 != nil {
				errCount++
				ctxLog.WithError(err1).Error("check isMatchByLogLevel failed")
				continue
			}
			isMatchedByLogAtAndLogLevel := isMatchedByLogAt && isMatchedByLogLevel
			prevLogEntry = LogEntry{
				LogAt:                       logLineInfo.GetTime(),
				LogLine:                     curLineBytes,
				LogLevel:                    logLevel,
				FileName:                    fileInfo.FileName,
				FileId:                      fileInfo.FileId,
				FileOffset:                  fileOffset,
				isMatchedByLogAtAndLogLevel: isMatchedByLogAtAndLogLevel,
				isMatched:                   isMatchedByLogAtAndLogLevel && isMatchedByKeyword,
			}
		} else {
			if isMatchedByKeyword {
				prevLogEntry.isMatched = prevLogEntry.isMatchedByLogAtAndLogLevel
			}
			prevLogEntry.LogLine = append(prevLogEntry.LogLine, curLineBytes...)
			prevLogEntry.FileOffset = fileOffset
		}
	}
	if prevLogEntry.isMatched {
		logQuery.SendLogEntry(prevLogEntry)
	}
	lastPos := &Position{
		FileId:     prevLogEntry.FileId,
		FileOffset: prevLogEntry.FileOffset,
	}
	if err := fdScanner.Err(); err != nil {
		ctxLog.WithError(err).Error("failed to scan")
		return lastPos, err
	}
	return lastPos, nil
}

func (l *LogQuerier) isMatchByLogAt(
	ctx context.Context,
	logLineInfo *message.Message,
	logQuery *LogQuery,
) (isMatched, isSkip bool, err error) {
	if logQuery.queryLogParams == nil {
		return false, false, nil
	}
	logAt := logLineInfo.GetTime()

	if logAt.Before(logQuery.queryLogParams.StartTime) {
		return false, false, nil
	}
	if logAt.After(logQuery.queryLogParams.EndTime) {
		return false, true, nil
	}
	return true, false, nil
}

func (l *LogQuerier) isMatchByLogLevel(
	ctx context.Context,
	logLevel string,
	logQuery *LogQuery,
) (isMatched bool, err error) {
	if logQuery.queryLogParams == nil {
		return false, nil
	}

	if logLevel == "" {
		return true, nil
	}

	if logLevel != "" && !isInArray(logLevel, logQuery.queryLogParams.LogLevel) {
		return false, nil
	}

	return true, nil
}

func (l *LogQuerier) isMatchByKeyword(
	ctx context.Context,
	logLineBytes []byte,
	logQuery *LogQuery,
) (isMatched bool, err error) {
	if logQuery.queryLogParams == nil {
		return false, nil
	}

	isMatchedKeyword := true
	if len(logQuery.keywordRegexps) != 0 {
		for _, keywordRegexp := range logQuery.keywordRegexps {
			if !keywordRegexp.Match(logLineBytes) {
				isMatchedKeyword = false
				break
			}
		}
	} else {
		for _, keyword := range logQuery.keywords {
			if !strings.Contains(string(logLineBytes), keyword) {
				isMatchedKeyword = false
				break
			}
		}
	}

	isMatchedExcludeKeyword := false
	if len(logQuery.excludeKeywordRegexps) != 0 {
		for _, excludeKeywordRegexp := range logQuery.excludeKeywordRegexps {
			if excludeKeywordRegexp.Match(logLineBytes) {
				isMatchedExcludeKeyword = true
				break
			}
		}
	} else {
		for _, excludeKeyword := range logQuery.excludeKeywords {
			if strings.Contains(string(logLineBytes), excludeKeyword) {
				isMatchedExcludeKeyword = true
				break
			}
		}
	}

	return isMatchedKeyword && !isMatchedExcludeKeyword, nil
}

// getMatchedFiles The files to be queried are matched according to the conditions
func (l *LogQuerier) getMatchedFiles(ctx context.Context, logQuery *LogQuery) ([]FileDetailInfo, error) {
	params := logQuery.queryLogParams
	ctxLog := log.WithContext(ctx).WithField("params", *params)
	filePatterns := getFilePattern(params.LogType, params.LogLevel, &logQuery.conf)
	var matchedFiles []FileDetailInfo
	for _, filePattern := range filePatterns {
		logAnalyzer := log_analyzer.GetLogAnalyzer(filePattern.LogAnalyzerCategory, params.LogType)
		if logAnalyzer == nil {
			ctxLog.Error("get LogInfoAnalyzer failed")
			continue
		}
		foundFiles, err := libFile.FindFilesByRegexAndTimeSpan(ctx, file.FindFilesParam{
			Dir:         filePattern.Dir,
			FileRegexps: filePattern.LogFilePatterns,
			MatchRegex:  matchString,
			StartTime:   params.StartTime,
			EndTime:     params.EndTime,
			GetFileTime: logAnalyzer.GetFileEndTime,
			MatchMTime:  matchMTime,
		})
		if err != nil {
			ctxLog.WithError(err).Error("FindFilesByRegexAndTimeSpan failed")
			continue
		}
		for _, foundFile := range foundFiles {
			fileTime, err1 := logAnalyzer.GetFileEndTime(foundFile)
			if err1 != nil {
				ctxLog.WithError(err1).Error("GetTimeFromFileName failed")
				continue
			}
			matchedFileName := filepath.Join(filePattern.Dir, foundFile.Name())
			ctxLog.WithField("matchedFileName", matchedFileName).Info("open matched file")
			fd, err1 := os.Open(matchedFileName)
			if err1 != nil {
				ctxLog.WithError(err1).Error("open matchedFile failed")
				continue
			}
			fileStatInfo, err1 := libFile.GetFileStatInfo(ctx, fd)
			if err1 != nil {
				ctxLog.WithError(err1).Error("GetFileStatInfo failed")
				continue
			}
			matchedFiles = append(matchedFiles, FileDetailInfo{
				LogAnalyzer: logAnalyzer,
				Dir:         filePattern.Dir,
				FileInfo:    foundFile,
				FileTime:    fileTime,
				FileDesc:    fd,
				FileId:      fileStatInfo.FileId(),
			})
		}
	}
	sort.Sort(ByFileTime(matchedFiles))

	newPos := 0
	if params.LastQueryFileId != 0 {
		for i := 0; i < len(matchedFiles); i++ {
			if matchedFiles[i].FileId == params.LastQueryFileId {
				newPos = i
				break
			}
		}
	}
	matchedFiles = matchedFiles[newPos:]
	if len(matchedFiles) != 0 {
		var offset int64
		if params.LastQueryFileOffset != 0 {
			offset = params.LastQueryFileOffset
		} else {
			newOffset, err := l.locateStartPosition(ctx, *matchedFiles[0].FileDesc, matchedFiles[0].LogAnalyzer, logQuery.queryLogParams.StartTime)
			if err != nil {
				ctxLog.WithError(err).Error("locateStartPosition failed")
				return nil, err
			}
			offset = newOffset
		}
		_, err := matchedFiles[0].FileDesc.Seek(offset, 0)
		if err != nil {
			ctxLog.WithError(err).Error("seek failed")
			return nil, err
		}
		matchedFiles[0].FileOffset = offset
	}
	return matchedFiles, nil
}

func (l *LogQuerier) getNextLineLogAt(
	reader io.Reader,
	logAnalyzer log_analyzer.LogAnalyzer,
) (*time.Time, error) {
	fdScanner := bufio.NewScanner(reader)
	maxCount := 100
	for i := 0; i < maxCount; i++ {
		fdScanner.Scan()
		lineBytes := fdScanner.Bytes()
		logLineInfo, isNewLine := logAnalyzer.ParseLine(string(lineBytes))
		if isNewLine {
			logAt := logLineInfo.GetTime()
			return &logAt, nil
		}
	}
	return nil, nil
}

func (l *LogQuerier) locateStartPosition(
	ctx context.Context,
	fd os.File,
	logAnalyzer log_analyzer.LogAnalyzer,
	at time.Time,
) (offset int64, err error) {
	ctxLog := log.WithContext(ctx)
	startPos := int64(0)
	endPos, err := fd.Seek(0, 2)
	if err != nil {
		ctxLog.WithError(err).Error("seek failed")
		return 0, err
	}

	for startPos < endPos && endPos-startPos > l.minPosGap {
		midPos := startPos + (endPos-startPos)>>1
		_, err1 := fd.Seek(midPos, 0)
		if err1 != nil {
			ctxLog.WithError(err).Error("seek midPos failed")
			return 0, err1
		}
		logAt, err1 := l.getNextLineLogAt(&fd, logAnalyzer)
		if err1 != nil {
			ctxLog.WithError(err).Error("getNextLineLogAt failed")
			return 0, err1
		}
		if logAt == nil {
			return 0, nil
		}
		if logAt.After(at) || logAt.Equal(at) {
			endPos = midPos
		} else if logAt.Before(at) {
			startPos = midPos + 1
		}
	}
	return startPos, nil
}
