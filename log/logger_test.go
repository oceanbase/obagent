package log

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func initlog() *logrus.Logger {
	return InitLogger(LoggerConfig{
		Level:      "debug",
		Filename:   "../tests/test.log",
		MaxSize:    10, // 10M
		MaxAge:     3,  // 3days
		MaxBackups: 3,
		LocalTime:  false,
		Compress:   false,
	})
}

func TestLogExample(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	_ = buf
	logger := initlog()

	// use logger
	logger.Debugf("debug-log-%d", 1)
	logger.WithField("field-key-1", "field-val-1").Infof("info-log-%d", 1)

	// with context, set traceId
	ctx := context.WithValue(context.Background(), TraceIdKey{}, "TRACE-ID")
	ctxlog := logger.WithContext(ctx)
	ctxlog.Debugf("debug-log-%d", 2)
	fieldlog := ctxlog.WithFields(map[string]interface{}{
		"field-key-2": "field-val-2",
		"field-key-3": "field-val-3",
	})
	// use the same field logger to avoid allocte new Entry
	fieldlog.Infof("info-log-%d", 2)
	fieldlog.Infof("info-log-%s", "2.1")

	// use logrus
	logrus.Debugf("debug-log-%d", 3)
	logrus.WithField("field-key-3", "field-val-3").Infof("info-log-%d", 3)
	fmt.Printf("%s", buf.Bytes())
}

func TestLogFile(t *testing.T) {
	initlog()

	// use logrus
	logrus.Debugf("debug-log-%d", 1)
	logrus.WithField("field-key-1", "field-val-1").Infof("info-log-%d", 1)
}

func TestLogDuration(t *testing.T) {
	initlog()

	ctx := context.WithValue(context.Background(), StartTimeKey, time.Now())
	log := logrus.WithContext(ctx)
	time.Sleep(time.Millisecond)
	log.Infof("test duration 1ms")
}

func TestFields(t *testing.T) {
	initlog()

	Fields("key1", 100, 2, 200.001, 3, "300").Infof("test fields")
	Fields().Infof("test 0 fields")
	Fields("1").Infof("test 1 fields")
	Fields("key1", 100, 2).Infof("test 3 fields")
}

func TestStdout(t *testing.T) {
	logger := logrus.StandardLogger()
	logger.SetOutput(os.Stdout)
	logrus.Infof("test stdout")
}

type errWriter struct{}

func (w *errWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write err")
}

func (w *errWriter) Close() error {
	return nil
}

func TestNoErrWriter(t *testing.T) {
	w := &noErrWriter{
		w: &errWriter{},
	}
	msg := "hello~"
	n, err := w.Write([]byte(msg))
	if err != nil {
		t.Error("should no error")
	}
	if n != len(msg) {
		t.Error("length wrong")
	}
}
