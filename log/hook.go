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
	"errors"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type CallerHook struct{}

func (hook *CallerHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 4)
	cnt := runtime.Callers(8, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !isIgnorePackages(name) {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data[logrus.FieldKeyFile] = getPackage(file)
			entry.Data[FieldKeyLine] = line
			entry.Data[logrus.FieldKeyFunc] = name[strings.LastIndex(name, ".")+1:]
			break
		}
	}
	return nil
}

func (hook *CallerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func isIgnorePackages(name string) bool {
	return strings.Contains(name, "github.com/sirupsen/logrus") ||
		strings.Contains(name, "github.com/go-kit/log") ||
		strings.Contains(name, "github.com/gin-gonic/gin")
}

type CostDurationHook struct{}

func (hook *CostDurationHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}
	startTime := entry.Context.Value(StartTimeKey)
	if startTime == nil {
		return nil
	}
	start, ok := startTime.(time.Time)
	if !ok {
		return errors.New("startTime is no time.Time")
	}

	duration := time.Now().Sub(start)
	entry.Data[FieldKeyDuration] = duration
	return nil
}

func (hook *CostDurationHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
