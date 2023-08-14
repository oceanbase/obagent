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

package sls

import (
	"fmt"
	"time"

	"github.com/aliyun/aliyun-log-go-sdk"
	"github.com/cenkalti/backoff"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/monitor/message"
)

type Config struct {
	AccessKeyID     string        `yaml:"accessKeyId"`
	AccessKeySecret string        `yaml:"accessKeySecret"`
	Endpoint        string        `yaml:"endpoint"`
	RequestTimeout  time.Duration `yaml:"requestTimeout"`
	ProjectName     string        `yaml:"projectName"`
	LogStoreName    string        `yaml:"logStoreName"`
	Topic           string        `yaml:"topic"`
	Source          string        `yaml:"source"`
	FieldMap        FieldMap      `yaml:"fieldMap"`
	Retry           Retry         `yaml:"retry"`
}

type FieldMap struct {
	Name   string            `yaml:"name"`
	Tags   map[string]string `yaml:"tags"`
	Fields map[string]string `yaml:"fields"`
}

type Retry struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
}

type SLSOutput struct {
	config       *Config
	client       sls.ClientInterface
	logStore     *sls.LogStore
	batch        []*message.Message
	done         chan struct{}
	retryBackoff backoff.BackOff
}

var DefaultConfig = Config{
	RequestTimeout: time.Second * 5,
	Retry: Retry{
		MaxInterval:     time.Minute,
		InitialInterval: time.Millisecond * 200,
		MaxElapsedTime:  time.Minute * 5,
	},
}

func NewSLSOutput(config *Config) *SLSOutput {
	sls.RetryOnServerErrorEnabled = false // disable sls internal retry
	client := &sls.Client{
		Endpoint:        config.Endpoint,
		AccessKeyID:     config.AccessKeyID,
		AccessKeySecret: config.AccessKeySecret,
		RequestTimeOut:  config.RequestTimeout,
	}
	retryBackoff := backoff.NewExponentialBackOff()
	retryBackoff.MaxInterval = config.Retry.MaxInterval
	retryBackoff.InitialInterval = config.Retry.InitialInterval
	retryBackoff.MaxElapsedTime = config.Retry.MaxElapsedTime
	retryBackoff.RandomizationFactor = 0.1
	return &SLSOutput{
		config:       config,
		client:       client,
		done:         make(chan struct{}),
		retryBackoff: retryBackoff,
	}
}

func (s *SLSOutput) Start(in <-chan []*message.Message) error {
	store, err := s.client.GetLogStore(s.config.ProjectName, s.config.LogStoreName)
	if err != nil {
		log.Errorf("get logStore failed, err: %s", err)
		return err
	}
	s.logStore = store
	for {
		select {
		case batch := <-in:
			logGroup := s.toLogGroup(batch)
			if len(logGroup.Logs) == 0 {
				continue
			}
			s.writeLogGroup(logGroup)
		case <-s.done:
			log.Info("SLS Output exited")
			return nil
		}
	}
}

func (s *SLSOutput) toLogGroup(batch []*message.Message) *sls.LogGroup {
	logs := make([]*sls.Log, 0, len(batch))
	for _, msg := range batch {
		logs = append(logs, s.toLog(msg))
	}
	return &sls.LogGroup{
		Topic:  proto.String(s.config.Topic),
		Source: proto.String(s.config.Source),
		Logs:   logs,
		//LogTags: s.constantTags,
	}
}

func (s *SLSOutput) toLog(msg *message.Message) *sls.Log {
	content := make([]*sls.LogContent, 0, len(msg.Tags())+len(msg.Fields()))
	fieldMap := s.config.FieldMap
	for _, fieldEntry := range msg.Fields() {
		if fieldName, ok := fieldMap.Fields[fieldEntry.Name]; ok {
			content = append(content, &sls.LogContent{
				Key:   proto.String(fieldName),
				Value: proto.String(fmt.Sprint(fieldEntry.Value)),
			})
		}
	}
	for _, tagEntry := range msg.Tags() {
		if fieldName, ok := fieldMap.Tags[tagEntry.Name]; ok {
			content = append(content, &sls.LogContent{
				Key:   proto.String(fieldName),
				Value: proto.String(fmt.Sprint(tagEntry.Value)),
			})
		}
	}
	if fieldMap.Name != "" {
		content = append(content, &sls.LogContent{
			Key:   proto.String(fieldMap.Name),
			Value: proto.String(msg.GetName()),
		})
	}
	return &sls.Log{
		Time:     proto.Uint32(uint32(msg.GetTime().Unix())),
		Contents: content,
	}
}

func (s *SLSOutput) writeLogGroup(logs *sls.LogGroup) {
	s.retry(func() error {
		err := s.logStore.PutLogs(logs)
		if err != nil {
			log.WithError(err).Warn("write log group failed")
		}
		return err
	})
}

func (s *SLSOutput) retry(fn func() error) {
	_ = backoff.Retry(func() error {
		err := fn()
		if err == nil {
			return nil
		}
		select {
		case <-s.done:
			return backoff.Permanent(err)
		default:
		}
		if canRetry(err) {
			return err
		}
		return backoff.Permanent(err)
	}, s.retryBackoff)
}

func canRetry(err error) bool {
	if err == nil {
		return false
	}
	if slsErr, ok := err.(*sls.Error); ok {
		if slsErr.Code == sls.WRITE_QUOTA_EXCEED {
			return true
		}
		return slsErr.HTTPCode >= 500
	} else if slsErr, ok := err.(*sls.BadResponseError); ok {
		return slsErr.HTTPCode >= 500
	} else {
		return false
	}
}

func (s *SLSOutput) Stop() {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("recover from panic: %v", err)
		}
	}()
	err := s.client.Close()
	if err != nil {
		log.Errorf("close SLS Client got error: %v", err)
	}
	close(s.done)
}
