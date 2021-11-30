package log

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const logAtLayout = "2006-01-02 15:04:05.000000"
const logTimeInFileNameLayout = "20060102150405"

type logFileInfo struct {
	fileDesc *os.File
	// 该文件已经写满了，并被重命名过
	isRenamed bool
}


type LogConfig struct {
        LogDir string `yaml:"logDir"`
        LogFileName string `yaml:"logFileName"`
}

type reportedError struct {
	ErrorCode  int
	ReportedAt time.Time
}

type ILogAnalyzer interface {
	isErrLog(logLine string) bool
	isMatchFilterRules(logAt time.Time, logLine string, filterRuleRegexps []*filterRuleInfo) bool
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

// isMatchFilterRules 是否匹配过滤规则
func (l *logAnalyzer) isMatchFilterRules(logAt time.Time, logLine string, filterRuleRegexps []*filterRuleInfo) bool {
	for _, rule := range filterRuleRegexps {
		matched := logAt.Before(rule.expireTime) && rule.reg.MatchString(logLine)
		if matched {
			log.WithFields(log.Fields{
				"rule":       rule.reg.String(),
				"logLine":    logLine,
				"logAt":      logAt,
				"expireTime": rule.expireTime,
			}).Info("match a filter rule")
			return true
		}
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

func compileFilterRules(filterRules []*FilterRule) (map[string][]*filterRuleInfo, error) {
	filterRuleRegexps := make(map[string][]*filterRuleInfo)
	if filterRules == nil {
		return filterRuleRegexps, nil
	}
	for _, rule := range filterRules {
		reg, err := regexp.Compile(rule.Keyword)
		if err != nil {
			return nil, err
		}
		ruleInfo := &filterRuleInfo{
			reg:        reg,
			expireTime: rule.ExpireTime,
		}
		if _, ok := filterRuleRegexps[rule.ServerType]; !ok {
			filterRuleRegexps[rule.ServerType] = []*filterRuleInfo{ruleInfo}
		} else {
			filterRuleRegexps[rule.ServerType] = append(filterRuleRegexps[rule.ServerType], ruleInfo)
		}
	}
	return filterRuleRegexps, nil
}

func processLogAlarmFiltersConf(logFilterRulesJsonContent string) ([]*FilterRule, error) {
	if logFilterRulesJsonContent == "" {
		return nil, nil
	}
	jsonStr := []byte(logFilterRulesJsonContent)
	var filterRules []*FilterRule
	err := json.Unmarshal(jsonStr, &filterRules)
	if err != nil {
		return nil, err
	}
	return filterRules, nil
}

func getTimeFromFileName(info os.FileInfo) (time.Time, error) {
	fileTime := info.ModTime()
	fileName := info.Name()
	lastDotIdx := strings.LastIndex(fileName, ".")
	timeStr := fileName[lastDotIdx+1:]
	if lastDotIdx == -1 || timeStr == "" {
		return fileTime, nil
	}
	parsedTime, err := time.ParseInLocation(logTimeInFileNameLayout, timeStr, time.Local)
	if err != nil {
		return fileTime, nil
	}
	return parsedTime, nil
}

// findFilesAndSortByMTime 获取路径下匹配正则的文件 & 文件大小（按照修改时间从旧到新排序）
func findFilesAndSortByMTime(ctx context.Context, dir, fileName string, start, end time.Time) ([]os.FileInfo, error) {
	ctxLog := log.WithContext(ctx).WithFields(log.Fields{
		"dir":      dir,
		"fileName": fileName,
		"start":    start,
		"end":      end,
	})
	mTime := -1 * end.Sub(start)
	matchedFiles, err := libFile.FindFilesByRegexAndMTime(ctx, dir, fileName, matchString, mTime, end, getTimeFromFileName)
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
	matched = strings.Contains(content, reg)
	return
}

type processQueue struct {
	queue []*logFileInfo
	mutex sync.Mutex
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

func getFileStatInfo(ctx context.Context, file *os.File) (*FileInfoEx, error) {
	ctxLog := log.WithContext(ctx)
	ret, err := FileInfo(file)
	if err != nil {
		ctxLog.WithError(err).Error("failed get file info")
		return nil, err
	}
	return ret, nil
}
