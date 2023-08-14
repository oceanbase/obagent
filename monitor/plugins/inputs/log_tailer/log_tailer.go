/*
 * Copyright (c) 2023 OceanBase
 * OBAgent is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package log_tailer

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/stat"
)

// LogTailer tails log for metric platform or others.
type LogTailer struct {
	conf        monagent.LogTailerConfig
	executors   []*LogTailerExecutor
	toBeStopped chan bool
}

// NewLogTailer creates a new instance of LogTailer
func NewLogTailer(conf monagent.LogTailerConfig) (*LogTailer, error) {
	if conf.ProcessQueueCapacity == 0 {
		conf.ProcessQueueCapacity = 10
	}
	return &LogTailer{
		conf:        conf,
		executors:   make([]*LogTailerExecutor, 0),
		toBeStopped: make(chan bool),
	}, nil
}

func (l *LogTailer) isStopped() bool {
	for _, executor := range l.executors {
		if !executor.isStopped() {
			return false
		}
	}
	return true
}

func (l *LogTailer) Start(out chan<- []*message.Message) (err error) {
	log.Info("start to run log tailer")
	err = l.Run(context.Background(), out)
	if err != nil {
		log.Error("failed to run log tailer")
		return err
	}
	return nil
}

func (l *LogTailer) Stop() {
	close(l.toBeStopped)
	// graceful stop
	i := 0
	for !l.isStopped() && i < 8000 {
		time.Sleep(time.Microsecond)
		i++
	}
	log.Infof("logTailer stopped, isStopped:%t", l.isStopped())
	stat.LogTailerCount.With(prometheus.Labels{stat.LogFileName: "all"}).Dec()
}

func (l *LogTailer) Run(ctx context.Context, out chan<- []*message.Message) error {
	ctxLog := log.WithContext(ctx)
	for _, tailConfig := range l.conf.TailConfigs {
		executor := NewLogTailerExecutor(tailConfig, l.conf.RecoveryConfig, l.toBeStopped, out)
		l.executors = append(l.executors, executor)
		err := executor.TailLog(ctx)
		if err != nil {
			ctxLog.WithField("tailConfig", tailConfig).WithError(err).Warn("failed to process log")
		}
	}

	return nil
}
