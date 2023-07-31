package log_analyzer

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/oceanbase/obagent/monitor/message"
)

const agentLogTimeFormat = "2006-01-02T15:04:05.99999-07:00"

type AgentLogAnalyzer struct {
	name string
}

var agentLogPattern = regexp.MustCompile(`^(?P<time>[\dT:.+-]+) (?P<level>[A-Z]+) \[(?P<pid>\d+),(?P<trace>\w*)\] caller=(?P<source>[\w./]+):(?P<line>\d+):(?P<func>\w+): (?P<content>.+)`)
var agentSubMatchIndex = make(map[string]int)

func init() {
	for i, name := range agentLogPattern.SubexpNames() {
		agentSubMatchIndex[name] = i
	}
}

func NewAgentLogAnalyzer(name string) LogAnalyzer {
	return &AgentLogAnalyzer{
		name: name,
	}
}

func (a *AgentLogAnalyzer) ParseLine(line string) (*message.Message, bool) {
	subMatch := agentLogPattern.FindStringSubmatch(line)
	if subMatch == nil {
		return nil, false
	}
	if agentSubMatchIndex["content"] >= len(subMatch) {
		return nil, false
	}
	t, err := time.ParseInLocation(agentLogTimeFormat, subMatch[agentSubMatchIndex["time"]], time.Local)
	if err != nil {
		return nil, false
	}
	msg := message.NewMessage(a.name, message.Log, t)
	msg.AddField("raw", line)
	msg.AddField("content", subMatch[agentSubMatchIndex["content"]])
	msg.AddTag("level", strings.ToLower(subMatch[agentSubMatchIndex["level"]]))
	msg.AddTag("pid", subMatch[agentSubMatchIndex["pid"]])
	msg.AddTag("source", subMatch[agentSubMatchIndex["source"]])
	trace := subMatch[agentSubMatchIndex["trace"]]
	if trace != "" {
		msg.AddTag("source", trace)
	}
	return msg, true
}

func (a *AgentLogAnalyzer) GetFileEndTime(info os.FileInfo) (time.Time, error) {
	return info.ModTime(), nil
}
