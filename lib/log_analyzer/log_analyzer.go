package log_analyzer

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/oceanbase/obagent/monitor/message"
)

type LogAnalyzer interface {
	ParseLine(line string) (msg *message.Message, isNewLine bool)
	GetFileEndTime(logFile os.FileInfo) (time.Time, error)
}

type LogAnalyzerFactory func(fileName string) LogAnalyzer

const (
	// TypeOb 结构化解析 observer 这类日志格式
	TypeOb = "ob"
	// TypeOb 轻量级解析 observer 这类日志格式（仅解析日志中时间戳、日志级别等）
	TypeObLight    = "ob_light"
	TypeAgent      = "agent"
	TypeAgentLight = "agent_light"
	TypeHost       = "host"
	TypeHostLight  = "host_light"
)

var analyzerFactories = map[string]LogAnalyzerFactory{
	TypeOb:         NewObLogAnalyzer,
	TypeObLight:    NewObLogLightAnalyzer,
	TypeAgent:      NewAgentLogAnalyzer,
	TypeAgentLight: NewAgentLogLightAnalyzer,
	TypeHost:       NewHostLogAnalyzer,
	TypeHostLight:  NewHostLogLightAnalyzer,
}

func GetLogAnalyzer(typeName string, fileName string) LogAnalyzer {
	factory := analyzerFactories[typeName]
	if factory == nil {
		panic("LogAnalyzerFactory " + typeName + " not exists")
	}
	return factory(fileName)
}

func ParseScanner(a LogAnalyzer, scanner *bufio.Scanner, msgConsumer func(*message.Message) bool) error {
	return ParseLines(a, func() (string, bool, error) {
		ok := scanner.Scan()
		if !ok {
			return "", false, scanner.Err()
		}
		return scanner.Text(), true, nil
	}, msgConsumer)
}

func ParseLines(a LogAnalyzer, lineProvider func() (string, bool, error), msgConsumer func(*message.Message) bool) error {
	var pendingLines []string
	var prevMsg *message.Message = nil
	for {
		line, ok, err := lineProvider()
		if !ok {
			if err != nil {
				return err
			}
			break
		}
		msg, isNewLine := a.ParseLine(line)
		if !isNewLine {
			pendingLines = append(pendingLines, line)
			continue
		}
		if prevMsg != nil {
			extra := strings.Join(pendingLines, "\n")
			prevMsg.AddField("extra", extra)
			pendingLines = nil
			ok1 := msgConsumer(prevMsg)
			if !ok1 {
				// 无法继续消费，那么退出循环
				break
			}
		}
		prevMsg = msg
	}
	if prevMsg != nil {
		extra := strings.Join(pendingLines, "\n")
		prevMsg.AddField("extra", extra)
		msgConsumer(prevMsg)
	}
	return nil
}

func ParseChan(a LogAnalyzer, lines <-chan string, messages chan<- *message.Message) error {
	defer close(messages)
	return ParseLines(a,
		func() (string, bool, error) {
			line, ok := <-lines
			return line, ok, nil
		},
		func(msg *message.Message) bool {
			messages <- msg
			return true
		},
	)
}

func ParseTimeFromFileName(fileName string, delimiter string, pattern string, def time.Time) time.Time {
	lastDotIdx := strings.LastIndex(fileName, delimiter)
	timeStr := fileName[lastDotIdx+1:]
	if lastDotIdx == -1 || timeStr == "" {
		return def
	}
	parsedTime, err := time.ParseInLocation(pattern, timeStr, time.Local)
	if err != nil {
		return def
	}
	return parsedTime
}
