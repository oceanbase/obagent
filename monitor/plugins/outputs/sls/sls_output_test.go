package sls

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"

	"github.com/oceanbase/obagent/monitor/message"
)

var testConfig = &Config{
	AccessKeyID:     "...",
	AccessKeySecret: "...",
	Endpoint:        "...",
	ProjectName:     "ocp-test",
	LogStoreName:    "ob_log",
	Topic:           "ocp_ob_log",
	Source:          "test",
	Retry: Retry{
		MaxInterval:     time.Millisecond * 500,
		InitialInterval: time.Millisecond * 100,
		MaxElapsedTime:  time.Second * 1,
	},
	FieldMap: FieldMap{
		Name: "name",
		Tags: map[string]string{
			"app":    "app",
			"level":  "level",
			"module": "module",
			"file":   "file",
			"func":   "func",
			"trace":  "trace",
			"thread": "thread",
			"code":   "code",
		},
		Fields: map[string]string{
			"content": "content",
			"error":   "error",
		},
	},
}

var messages = []*message.Message{
	message.NewMessage("test", message.Log, time.Now()).
		AddTag("level", "info").
		AddTag("app", "ob").
		AddTag("file", "a.cpp").
		AddTag("module", "SQL").
		AddTag("func", "test_func").
		AddTag("trace", "Y0000000000").
		AddTag("thread", "100000").
		AddTag("code", "1423").
		AddField("content", "this is a test log 1"),
	message.NewMessage("test", message.Log, time.Now()).
		AddTag("level", "info").
		AddTag("app", "ob").
		AddTag("file", "a.cpp").
		AddTag("module", "SQL").
		AddTag("func", "test_func").
		AddTag("trace", "Y0000000001").
		AddTag("thread", "100001").
		AddTag("code", "1423").
		AddField("content", "this is a test log 2"),
	message.NewMessage("test", message.Log, time.Now()).
		AddTag("level", "info").
		AddTag("app", "ob").
		AddTag("file", "a.cpp").
		AddTag("module", "SQL").
		AddTag("func", "test_func").
		AddTag("trace", "Y0000000002").
		AddTag("thread", "100002").
		AddTag("code", "1423").
		AddField("content", "this is a test log 3"),
}

var testConfig2 = &Config{
	//AccessKeyID:     "...",
	//AccessKeySecret: "...",
	//Endpoint:        "...",
	ProjectName:  "ocp-test",
	LogStoreName: "log2",
	Topic:        "ocp_ob_log",
	Source:       "test",
	Retry: Retry{
		MaxInterval:     time.Millisecond * 500,
		InitialInterval: time.Millisecond * 100,
		MaxElapsedTime:  time.Second * 1,
	},
	FieldMap: FieldMap{
		Name: "name",
		Tags: map[string]string{
			"app":   "app",
			"level": "level",
		},
		Fields: map[string]string{
			"content": "content",
			"tags":    "tags",
			"error":   "error",
		},
	},
}

var messages2 = []*message.Message{
	message.NewMessage("test", message.Log, time.Now()).
		AddTag("level", "info").
		AddTag("app", "ob").
		AddField("tags", "file:a.cpp\nmodule:SQL\nfunc:test_func\ntrace:Y0000000000\nthread:100000\ncode:1423").
		AddField("content", "this is a test log 1"),
	message.NewMessage("test", message.Log, time.Now()).
		AddTag("level", "info").
		AddTag("app", "ob").
		AddField("tags", "file:a.cpp\nmodule:SQL\nfunc:test_func\ntrace:Y0000000000\nthread:100000\ncode:1423").
		AddField("content", "this is a test log 2"),
	message.NewMessage("test", message.Log, time.Now()).
		AddTag("level", "info").
		AddTag("app", "ob").
		AddField("tags", "file:a.cpp\nmodule:SQL\nfunc:test_func\ntrace:Y0000000000\nthread:100000\ncode:1423").
		AddField("content", "this is a test log 3"),
}

func Test_Stop(t *testing.T) {
	output := NewSLSOutput(testConfig)
	output.Stop()
	output.Stop()
}

func Test_toLog(t *testing.T) {
	output := NewSLSOutput(testConfig)
	msg := messages[0]
	log := output.toLog(msg)
	if log.GetTime() != uint32(msg.GetTime().Unix()) {
		t.Errorf("time wrong")
	}
	contents := log.GetContents()
	if len(contents) != len(msg.Tags())+len(msg.Fields())+1 {
		t.Errorf("wrong content count")
	}
}

func Test_toLogGroup(t *testing.T) {
	output := NewSLSOutput(testConfig)
	group := output.toLogGroup(messages)
	if len(group.Logs) != len(messages) {
		t.Errorf("log count wrong")
	}
	if group.GetTopic() != testConfig.Topic {
		t.Errorf("topic wrong")
	}
	if group.GetSource() != testConfig.Source {
		t.Errorf("source wrong")
	}
}

func Test_canRetry(t *testing.T) {
	if canRetry(nil) {
		t.Errorf("nil can not retry")
	}
	if canRetry(errors.New("xxx")) {
		t.Errorf("none http error can not retry")
	}
	if canRetry(&sls.Error{HTTPCode: -1}) {
		t.Errorf("client error can not retry")
	}
	if canRetry(&sls.Error{HTTPCode: 401}) {
		t.Errorf("client error can not retry")
	}
	if !canRetry(&sls.Error{HTTPCode: 500}) {
		t.Errorf("server error can retry")
	}
	if canRetry(&sls.Error{HTTPCode: 499}) {
		t.Errorf("client error can not retry")
	}
	if canRetry(&sls.BadResponseError{HTTPCode: 400}) {
		t.Errorf("client error can not retry")
	}
	if !canRetry(&sls.BadResponseError{HTTPCode: 500}) {
		t.Errorf("server error can retry")
	}
}

func Test_retry(t *testing.T) {
	output := NewSLSOutput(testConfig)
	i := 0
	output.retry(func() error {
		i++
		return nil
	})
	if i != 1 {
		t.Errorf("retry count wrong")
	}
	i = 0
	output.retry(func() error {
		i++
		return errors.New("xx")
	})
	if i != 1 {
		t.Errorf("retry count wrong")
	}
	i = 0
	output.retry(func() error {
		i++
		return &sls.Error{HTTPCode: 503}
	})
	println(i)
	if i < 2 {
		t.Errorf("retry count wrong")
	}
}

func TestSLSOutput(t *testing.T) {
	t.Skip()
	output := NewSLSOutput(testConfig2)
	in := make(chan []*message.Message, 100)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := output.Start(in)
		fmt.Println(err)
		wg.Done()
	}()
	in <- messages2

	fmt.Println("stop...")
	output.Stop()
	fmt.Println("wait done...")
	wg.Wait()
}
