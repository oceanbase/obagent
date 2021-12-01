package log

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const logAtLayout = "2006-01-02 15:04:05.000000"
const logTimeInFileNameLayout = "20060102150405"

type logFileInfo struct {
	fileDesc *os.File
	// 该文件已经写满了，并被重命名过
	isRenamed bool
}

type LogConfig struct {
	LogDir      string `yaml:"logDir"`
	LogFileName string `yaml:"logFileName"`
}

type reportedError struct {
	ErrorCode  int
	ReportedAt time.Time
}

type ILogAnalyzer interface {
	isErrLog(logLine string) bool
	getErrCode(logLine string) (int, error)
	getLogAt(logLine string) (time.Time, error)
}

type logAnalyzer struct {
	logAtRegexp   *regexp.Regexp
	errCodeRegexp *regexp.Regexp
}

func NewLogAnalyzer() *logAnalyzer {
	return &logAnalyzer{
		logAtRegexp:   regexp.MustCompile(`^\[\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d\.\d\d\d\d\d\d\]`),
		errCodeRegexp: regexp.MustCompile(`ret=-\d+`),
	}
}

// isErrLog 检查是否 error 类型的 log（后面看需求是否需要扩展为其他类型日志）
func (l *logAnalyzer) isErrLog(logLine string) bool {
	// example: [2020-08-07 05:55:44.377075] ERROR  [RS] ob_server_table_operator.cpp:376 [151575][4][Y0-0000000000000000] [lt = 5] [dc =0] svr_status(svr_status = "active", display_status =1)
	if len(logLine) > 34 {
		return "ERROR" == logLine[29:34]
	}

	return false
}

// getErrCode 获取日志中的错误码
func (l *logAnalyzer) getErrCode(logLine string) (int, error) {
	matchedErrCodes := l.errCodeRegexp.FindAllString(logLine, -1)
	matchedErrCodesLen := len(matchedErrCodes)
	if matchedErrCodesLen > 0 {
		lastErrCodeStr := matchedErrCodes[matchedErrCodesLen-1]
		// 匹配的格式为 ret=-\d+，数字从下标位置 5 开始
		if len(lastErrCodeStr) >= 5 {
			errCode, err := strconv.Atoi(lastErrCodeStr[5:])
			if err != nil {
				return -1, err
			}
			return errCode, nil
		}
	} else if strings.Contains(logLine, "clog disk is almost full") {
		return 4264, nil

	} else if strings.Contains(logLine, "partition table update task cost too much time to execute") {
		return 4015, nil
	}

	return -1, nil
}

func (l *logAnalyzer) getLogAt(logLine string) (time.Time, error) {
	timeStr := strings.TrimRight(strings.TrimLeft(l.logAtRegexp.FindString(logLine), "["), "]")
	logAt, err := time.ParseInLocation(logAtLayout, timeStr, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return logAt, nil
}

func matchString(reg string, content string) (matched bool, err error) {
	matched = strings.Contains(content, reg)
	return
}

type processQueue struct {
	queue []*logFileInfo
	mutex sync.Mutex
}

func (p *processQueue) getQueueLen() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()
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

func (p *processQueue) pushBack(info *logFileInfo) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.queue = append(p.queue, info)
}

func (p *processQueue) setHeadIsRenameTrue() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.getQueueLen() == 0 {
		return
	}

	p.queue[0].isRenamed = true
}

func (p *processQueue) getHeadIsRenamed() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.getQueueLen() == 0 {
		return false
	}

	return p.queue[0].isRenamed
}
