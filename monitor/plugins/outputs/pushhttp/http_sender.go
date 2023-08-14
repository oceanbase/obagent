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

package pushhttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/trace"
	"github.com/oceanbase/obagent/stat"
	log "github.com/sirupsen/logrus"
)

type Sender interface {
	Send(ctx context.Context, data []byte) error
	Close()
}

type HttpSenderConfig struct {
	TargetAddress         string        `yaml:"targetAddress"`
	ProxyAddress          string        `yaml:"proxyAddress"`
	APIUrl                string        `yaml:"apiUrl"`
	HttpMethod            string        `yaml:"httpMethod"`
	BasicAuthEnabled      bool          `yaml:"basicAuthEnabled"`
	Username              string        `yaml:"username"`
	Password              string        `yaml:"password"`
	Timeout               time.Duration `yaml:"timeout"`
	ContentType           string        `yaml:"contentType"`
	Headers               []string      `yaml:"headers"`
	AcceptedResponseCodes []int         `yaml:"acceptedResponseCodes"`
	retryTaskCount        int           `yaml:"-"`
	retryTimes            int           `yaml:"-"`
	// use default value if not set
	MaxIdleConns          int           `yaml:"maxIdleConns"`
	MaxConnsPerHost       int           `yaml:"maxConnsPerHost"`
	MaxIdleConnsPerHost   int           `yaml:"maxIdleConnsPerHost"`
	KeepAliveDuration     time.Duration `yaml:"keepAliveDuration"`
	IdleConnTimeout       time.Duration `yaml:"idleConnTimeout"`
	ResponseHeaderTimeout time.Duration `yaml:"responseHeaderTimeout"`
	ExpectContinueTimeout time.Duration `yaml:"expectContinueTimeout"`
}

type HttpSender struct {
	HttpSenderConfig

	client *http.Client

	retryTasks    chan retryTask
	stopped       chan bool
	taskWaitGroup sync.WaitGroup

	reservedIdleTaskCount int
	acceptedResponseCodes map[int]bool
}

const (
	DefaultMaxIdleConns          = 16
	DefaultMaxConnsPerHost       = 16
	DefaultMaxIdleConnsPerHost   = 16
	DefaultKeepAliveDuration     = time.Minute * 10
	DefaultIdleConnTimeout       = time.Second * 120
	DefaultResponseHeaderTimeout = time.Second * 2
	DefaultExpectContinueTimeout = time.Second * 2
)

type retryTask struct {
	retryTimes int
	createTime time.Time
	data       []byte
}

func (t retryTask) canBeDiscard(duration time.Duration) bool {
	return t.retryTimes > 0 && t.createTime.Add(duration).Before(time.Now())
}

func NewHttpSender(config HttpSenderConfig) Sender {
	s := &HttpSender{
		HttpSenderConfig:      config,
		reservedIdleTaskCount: 8,
		stopped:               make(chan bool),
		acceptedResponseCodes: make(map[int]bool),
	}
	if s.HttpSenderConfig.MaxIdleConns == 0 {
		s.HttpSenderConfig.MaxIdleConns = DefaultMaxIdleConns
	}
	if s.HttpSenderConfig.MaxIdleConnsPerHost == 0 {
		s.HttpSenderConfig.MaxIdleConnsPerHost = DefaultMaxIdleConnsPerHost
	}
	if s.HttpSenderConfig.MaxConnsPerHost == 0 {
		s.HttpSenderConfig.MaxConnsPerHost = DefaultMaxConnsPerHost
	}
	if s.HttpSenderConfig.KeepAliveDuration == 0 {
		s.HttpSenderConfig.KeepAliveDuration = DefaultKeepAliveDuration
	}
	if s.HttpSenderConfig.IdleConnTimeout == 0 {
		s.HttpSenderConfig.IdleConnTimeout = DefaultIdleConnTimeout
	}
	if s.HttpSenderConfig.ResponseHeaderTimeout == 0 {
		s.HttpSenderConfig.ResponseHeaderTimeout = DefaultResponseHeaderTimeout
	}
	if s.HttpSenderConfig.ExpectContinueTimeout == 0 {
		s.HttpSenderConfig.ExpectContinueTimeout = DefaultExpectContinueTimeout
	}

	for _, code := range config.AcceptedResponseCodes {
		s.acceptedResponseCodes[code] = true
	}

	s.client = s.buildHttpClient()
	s.retryTasks = make(chan retryTask, s.retryTaskCount)

	s.taskWaitGroup.Add(2)
	go s.startRetryTask()
	go s.discartRetryTask()

	return s
}

func (s *HttpSender) Close() {
	log.Info("http sender begin close")
	close(s.stopped)
	s.taskWaitGroup.Wait()
	log.Info("http sender closed")
}

func (s *HttpSender) Send(ctx context.Context, data []byte) error {
	stat.HttpOutputPushMetricsBytesTotal.With(prometheus.Labels{
		stat.HttpApiPath: s.APIUrl,
	}).Add(float64((len(data))))

	stat.HttpOutputPushTotal.With(prometheus.Labels{
		stat.HttpApiPath: s.APIUrl,
	}).Inc()

	err := s.send(ctx, bytes.NewReader(data))
	if err != nil {
		log.WithContext(ctx).Errorf("push metrics failed, err:%+v", err)
		if s.retryTimes > 0 {
			log.WithContext(ctx).Infof("add failed push task to retry tasks")
			select {
			case <-s.stopped:
			case s.retryTasks <- retryTask{
				data:       data,
				createTime: time.Now(),
			}:
			}

		}
	}
	return nil
}

func (s *HttpSender) send(ctx context.Context, reader io.Reader) error {
	startTime := time.Now()
	log.WithContext(ctx).Debugf("push metrics to %s", s.address())
	req, err := http.NewRequest(s.HttpMethod, s.address(), reader)
	if err != nil {
		return err
	}
	if s.BasicAuthEnabled {
		req.SetBasicAuth(s.Username, s.Password)
	}
	// set headers
	req.Header.Set("Content-Type", s.ContentType)
	for _, it := range s.Headers {
		headers := strings.SplitN(it, ":", 2)
		if len(headers) != 2 {
			return errors.Errorf("header %s is invalid", it)
		}
		req.Header.Add(headers[0], headers[1])
	}

	resp, err := s.client.Do(req)
	if err != nil {
		stat.HttpOutputSendFailedCount.With(prometheus.Labels{stat.HttpApiPath: s.APIUrl}).Inc()
		return err
	}
	defer resp.Body.Close()

	// stats
	stat.HttpOutputSendMillisecondsSummary.With(prometheus.Labels{
		stat.HttpMethod:  s.HttpMethod,
		stat.HttpStatus:  strconv.FormatInt(int64(resp.StatusCode), 10),
		stat.HttpApiPath: s.APIUrl,
	}).Observe(float64(time.Now().Sub(startTime)) / float64(time.Millisecond))

	if !s.acceptedResponseCodes[resp.StatusCode] {
		body, _ := io.ReadAll(resp.Body)
		return errors.Errorf("%s error code: %d, body: %s", s.address(), resp.StatusCode, body)
	}
	return nil
}

func (s *HttpSender) address() string {
	addr := s.TargetAddress
	if !strings.HasPrefix(s.TargetAddress, "http") {
		addr = "http://" + addr
	}
	return fmt.Sprintf("%s%s", addr, s.APIUrl)
}

func (s *HttpSender) buildHttpClient() *http.Client {
	var proxy func(*http.Request) (*url.URL, error) = nil
	if s.ProxyAddress != "" {
		proxy = func(_ *http.Request) (*url.URL, error) {
			addr := s.ProxyAddress
			if !strings.HasPrefix(s.ProxyAddress, "http") {
				addr = "http://" + s.ProxyAddress
			}
			return url.Parse(addr)
		}
	}
	transport := &http.Transport{
		Proxy: proxy,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: s.KeepAliveDuration,
			DualStack: true,
		}).DialContext,
		DisableKeepAlives:     false,
		MaxIdleConns:          s.MaxIdleConns,
		MaxConnsPerHost:       s.MaxConnsPerHost,
		MaxIdleConnsPerHost:   s.MaxIdleConnsPerHost,
		IdleConnTimeout:       s.IdleConnTimeout,
		ResponseHeaderTimeout: s.ResponseHeaderTimeout,
		ExpectContinueTimeout: s.ExpectContinueTimeout,
	}
	client := &http.Client{Transport: transport, Timeout: s.Timeout}
	return client
}

func (s *HttpSender) startRetryTask() {
	defer s.taskWaitGroup.Done()
	for {
		select {
		case _, open := <-s.stopped:
			if !open {
				log.Info("retryTask stopped")
				return
			}
		case task, open := <-s.retryTasks:
			if !open {
				return
			}
			ctx := trace.ContextWithRandomTraceId()
			task.retryTimes++
			stat.HttpOutputSendRetryCount.With(prometheus.Labels{stat.HttpApiPath: s.APIUrl}).Inc()
			err := s.send(ctx, bytes.NewReader(task.data))
			if err != nil {
				log.WithContext(ctx).Warnf("retry to push metrics failed, retry times: %d, err: %s", task.retryTimes, err)
			}
			if err != nil && task.retryTimes < s.retryTimes && len(s.retryTasks) < s.retryTaskCount {
				select {
				case <-s.stopped:
				case s.retryTasks <- task:
				}
			}
		}
	}
}

// To ensure that the new retry task can be executed, discard the existing retry task. The following strategies are used:
// The number of retry times is greater than 0 (retry still failed),
// and the creation time is earlier than two timeout periods (wait time exceeds a certain period).
// Until the retry queue length can accommodate the number of idle
func (s *HttpSender) discartRetryTask() {
	defer s.taskWaitGroup.Done()
	ticker := time.NewTicker(time.Second)
	discartRetryTimeout := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case _, open := <-s.stopped:
			if !open {
				log.Info("discartRetryTask stopped")
				return
			}
		case task, open := <-s.retryTasks:
			if !open {
				return
			}
			stat.HttpOutputRetryTaskSize.With(prometheus.Labels{stat.HttpApiPath: s.APIUrl}).Set(float64(len(s.retryTasks) + 1))
			if len(s.retryTasks) <= s.retryTaskCount-s.reservedIdleTaskCount {
				<-ticker.C
				continue
			}
			if !task.canBeDiscard(s.Timeout * 2) {
				select {
				case <-s.stopped:
				case <-discartRetryTimeout.C:
				case s.retryTasks <- task:
				}
			} else {
				log.Warnf("retry task is discarded, create at %+v, retry times %d", task.createTime, task.retryTimes)
			}
		}
	}
}
