package log_analyzer

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/oceanbase/obagent/monitor/message"
)

var obLogRegexp = regexp.MustCompile(`^\[(?P<time>[\d:. -]+)\]\s+(?:(?P<level>[A-Z_]+)\s+)?(?:\[(?P<module>[A-Z._]+)\](:\d*)?\s+)?(?:(?P<func>\S+) \((?P<file1>[\w./]+):(?P<line1>\d+)\)\s+|(?P<file2>\w+\.\w+):(?P<line2>\d+)\s+)?\[(?P<tid>\d+)\](?:\[\d+\]|\[(?P<thread>\w+)\]\[T(?P<tenant>\d+)\])?\[(?P<obLogTrace>[\w-]+)\]\s+(?:\[[\w=]+\]\s+)*(?P<content>.+)`)
var obErrCodeRegexp = regexp.MustCompile(`ret=-(\d+)`)

const obTimeLayout = "2006-01-02 15:04:05.000000"
const obLogFileTimePattern = "20060102150405"

var obSubMatchIndex = make(map[string]int)

func init() {
	for i, name := range obLogRegexp.SubexpNames() {
		obSubMatchIndex[name] = i
	}
}

type ObLogAnalyzer struct {
	fileName string
}

func NewObLogAnalyzer(fileName string) LogAnalyzer {
	return &ObLogAnalyzer{
		fileName: fileName,
	}
}

func (a *ObLogAnalyzer) ParseLine(line string) (*message.Message, bool) {
	subMatch := obLogRegexp.FindStringSubmatch(line)
	if subMatch == nil {
		return nil, false
	}
	t, err := time.ParseInLocation(obTimeLayout, subMatch[obSubMatchIndex["time"]], time.Local)
	if err != nil {
		return nil, false
	}
	msg := message.NewMessage(a.fileName, message.Log, t)
	msg.AddField("raw", line)
	for _, name := range obLogRegexp.SubexpNames() {
		if name == "" {
			continue
		}
		i := obSubMatchIndex[name]
		if i >= len(subMatch) {
			continue
		}
		value := subMatch[i]
		if value == "" {
			continue
		}
		switch name {
		case "time", "line1", "line2":
			continue
		case "level":
			msg.AddTag("level", strings.ToLower(value))
		case "content":
			msg.AddField("content", value)
			errCodeMatch := obErrCodeRegexp.FindStringSubmatch(value)
			if len(errCodeMatch) > 1 {
				msg.AddTag("errCode", errCodeMatch[1])
			}
		default:
			if name == "file1" || name == "file2" {
				name = "source"
			}
			msg.AddTag(name, value)
		}
	}
	return msg, true
}

func (a *ObLogAnalyzer) GetFileEndTime(info os.FileInfo) (time.Time, error) {
	return ParseTimeFromFileName(info.Name(), ".", obLogFileTimePattern, info.ModTime()), nil
}
