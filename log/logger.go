/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	StartTimeKey = "startTime"
)

const defaultTimestampFormat = "2006-01-02T15:04:05.99999-07:00"

var textFormatter = &TextFormatter{
	TimestampFormat:        defaultTimestampFormat, // log timestamp format
	FullTimestamp:          true,
	DisableLevelTruncation: true,
	FieldMap: map[string]string{
		"WARNING": "WARN", // log level string, use WARN
	},
	// log caller, filename:line callFunction
	CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
		filename := getPackage(frame.File)
		name := frame.Function
		idx := strings.LastIndex(name, ".")
		return name[idx+1:], fmt.Sprintf("%s:%d", filename, frame.Line)
	},
}

/**
 * wrap a writer to ignore error to avoid bad write cause logrus logger always print error messages.
 */
type noErrWriter struct {
	o sync.Once
	w io.WriteCloser
}

func (w *noErrWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if err != nil {
		w.o.Do(func() {
			// only print error message once
			_, _ = fmt.Fprintf(os.Stderr, "write log failed %v\n", err)
		})
		return len(p), nil
	}
	return
}

func (w *noErrWriter) Close() error {
	return w.w.Close()
}

type LoggerConfig struct {
	Level      string `yaml:"level"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxsize"`
	MaxAge     int    `yaml:"maxage"`
	MaxBackups int    `yaml:"maxbackups"`
	LocalTime  bool   `yaml:"localtime"`
	Compress   bool   `yaml:"compress"`
}

func InitLogger(config LoggerConfig) *logrus.Logger {
	logger := logrus.StandardLogger()
	if curOut, ok := logger.Out.(*noErrWriter); ok {
		if l, ok := curOut.w.(*lumberjack.Logger); ok {
			l.Filename = config.Filename
			l.MaxSize = config.MaxSize
			l.MaxBackups = config.MaxBackups
			l.MaxAge = config.MaxAge
			l.Compress = config.Compress
			_ = l.Close()
		}
	} else {
		writer := &lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
			LocalTime:  true,
		}
		logger.SetOutput(&noErrWriter{w: writer})
		// log format
		logger.SetFormatter(textFormatter)

		// use CallerHook, not ReportCaller
		logger.SetReportCaller(false)

		// log hook
		// cost duration hook
		logger.AddHook(new(CostDurationHook))
		// caller hook
		logger.AddHook(new(CallerHook))
	}
	// log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		panic(fmt.Sprintf("parse log level: %+v", err))
	}
	logger.SetLevel(level)
	return logger
}
