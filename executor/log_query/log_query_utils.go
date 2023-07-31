package log_query

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/oceanbase/obagent/config/mgragent"
)

const allLevel string = "ALL"

func getFilePattern(logType string, logLevels []string, logQueryConf *mgragent.LogQueryConfig) (filePatterns []DirAndFilePattern) {
	var logTypeQueryConfig mgragent.LogTypeQueryConfig
	for _, typeQueryConfig := range logQueryConf.LogTypeQueryConfigs {
		if typeQueryConfig.LogType == logType {
			logTypeQueryConfig = typeQueryConfig
			break
		}
	}
	matchedLogFilePatterns := make(map[string]DirAndFilePattern)
	for _, levelAndFilePattern := range logTypeQueryConfig.LogLevelAndFilePatterns {
		if !isInArray(levelAndFilePattern.LogLevel, logLevels) &&
			levelAndFilePattern.LogLevel != allLevel {
			continue
		}
		if !logTypeQueryConfig.IsOverrideByPriority {
			matchedLogFilePatterns[buildFilePatternStr(levelAndFilePattern.FilePatterns)] = DirAndFilePattern{
				Dir:                 levelAndFilePattern.Dir,
				LogFilePatterns:     levelAndFilePattern.FilePatterns,
				LogAnalyzerCategory: levelAndFilePattern.LogParserCategory,
			}
		} else {
			filePatterns = []DirAndFilePattern{
				{
					Dir:                 levelAndFilePattern.Dir,
					LogFilePatterns:     levelAndFilePattern.FilePatterns,
					LogAnalyzerCategory: levelAndFilePattern.LogParserCategory,
				},
			}
		}
	}
	if len(matchedLogFilePatterns) != 0 {
		for _, filePattern := range matchedLogFilePatterns {
			filePatterns = append(filePatterns, filePattern)
		}
	}

	return
}

func isInArray(str string, arr []string) bool {
	for _, val := range arr {
		if str == val {
			return true
		}
	}
	return false
}

func matchString(regs []string, content string) (matched bool, err error) {
	for _, reg := range regs {
		matchContent, err1 := filepath.Match(reg, content)
		if err1 != nil {
			return false, err1
		}
		if matchContent {
			return true, nil
		}
	}
	return
}

// ByFileTime implements sort.Interface for []FileDetailInfo based on FileTime.
type ByFileTime []FileDetailInfo

func (a ByFileTime) Len() int           { return len(a) }
func (a ByFileTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFileTime) Less(i, j int) bool { return a[i].FileTime.Before(a[j].FileTime) }

func matchMTime(mTime, startTime, endTime time.Time) (matched bool, err error) {
	return mTime.After(startTime), nil
}

func dropCRLF(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[0 : len(data)-1]
		if len(data) > 0 && data[len(data)-1] == '\r' {
			data = data[0 : len(data)-1]
		}
	}
	return data
}

func buildFilePatternStr(filePattern []string) string {
	return strings.Join(filePattern, ",")
}
