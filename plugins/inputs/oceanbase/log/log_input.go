package log

import (
	"bufio"
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/metric"
)

const sampleConfig = `
expireTime: 300s
collectDelay: 1s
LogServiceConfig:
  rootservice:
    excludeRegexes:
      - hello
      - world
    logConfig:
      logDir: /home/admin/oceanbase/log
      logFileName: rootservice.log.wf
  election:
    excludeRegexes:
      - hello
      - world
    logConfig:
      logDir: /home/admin/oceanbase/log
      logFileName: election.log.wf
  observer:
    excludeRegexes:
      - hello
      - world
    logConfig:
      logDir: /home/admin/oceanbase/log
      logFileName: observer.log.wf
`

const description = `
collect ob error logs and filter by keywords
`

type ServiceType string

const (
	RootService ServiceType = "rootservice"
	Observer    ServiceType = "observer"
	Election    ServiceType = "election"
)

type LogCollectConfig struct {
	LogConfig      *LogConfig `yaml:"logConfig"`
	ExcludeRegexes []string   `yaml:"excludeRegexes"`
}

type Config struct {
	LogServiceConfig map[ServiceType]*LogCollectConfig `yaml:"logServiceConfig"`
	CollectDelay     time.Duration                     `yaml:"collectDelay"`
	ExpireTime       time.Duration                     `yaml:"expireTime"`
}

type ErrorLogInput struct {
	config           *Config
	logAnalyzer      ILogAnalyzer
	logProcessQueue  map[ServiceType]*processQueue
	ctx              context.Context
	cancel           context.CancelFunc
	metricBufferChan chan []metric.Metric
}

func (e *ErrorLogInput) SampleConfig() string {
	return sampleConfig
}

func (e *ErrorLogInput) Description() string {
	return description
}

func (e *ErrorLogInput) Init(config map[string]interface{}) error {

	var pluginConfig Config
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "error log input encode config")
	}
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "error log input decode config")
	}
	e.config = &pluginConfig

	e.logAnalyzer = NewLogAnalyzer()
	e.logProcessQueue = make(map[ServiceType]*processQueue)

	// start go routine to add log file to logProcessQueue
	go e.watchFile()

	for service := range e.config.LogServiceConfig {
		go e.doCollect(service)
	}

	log.Info("error log input init with config", e.config)

	return nil
}

func (e *ErrorLogInput) doCollect(service ServiceType) {
	for {
		select {
		case <-e.ctx.Done():
			log.Infof("received exit signal, stop collect routine of service %s", service)
			q, found := e.logProcessQueue[service]
			if found {
				for q.getQueueLen() > 0 {
					fd := q.getHead().fileDesc
					err := fd.Close()
					if err != nil {
						log.Errorf("close log file of service %s %s failed %v", service, fd.Name(), err)
					}
					q.popHead()
				}
			}
		default:
			e.collectErrorLogs(service)
			time.Sleep(e.config.CollectDelay)
		}
	}
}

func (e *ErrorLogInput) collectErrorLogs(service ServiceType) {
	q, found := e.logProcessQueue[service]
	if !found {
		log.Warnf("service %s has no process queue", service)
	} else {
		if q.getQueueLen() == 0 {
			log.Warnf("service %s has no process queue", service)
		} else {
			// read head of queue
			logFile := q.getHead()
			fdScanner := bufio.NewScanner(logFile.fileDesc)
			logMetrics := make([]metric.Metric, 0, 8)
			for fdScanner.Scan() {
				line := fdScanner.Text()
				if line == "" || len(line) == 0 {
					continue
				} else {
					logMetric := e.processLogLine(service, line)
					if logMetric != nil {
						logMetrics = append(logMetrics, logMetric)
					}
				}
			}
			if len(logMetrics) > 0 {
				e.metricBufferChan <- logMetrics
			}
			if logFile.isRenamed {
				logFile.fileDesc.Close()
				q.popHead()
			}
		}
	}
}

func (e *ErrorLogInput) processLogLine(service ServiceType, line string) metric.Metric {
	if !e.logAnalyzer.isErrLog(line) {
		return nil
	}
	logAt, err := e.logAnalyzer.getLogAt(line)
	if err != nil {
		log.Warnf("parse log time failed %s ", line)
	}
	if logAt.Add(e.config.ExpireTime).Before(time.Now()) {
		log.Debugf("log expired, just skip, %s", line)
		return nil
	}
	errCode, err := e.logAnalyzer.getErrCode(line)
	if err != nil {
		log.Warnf("parse log err code failed %s ", line)
	}
	if e.isFiltered(service, line) {
		log.Debugf("log is filtered, %s", line)
		return nil
	}
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	fields["log_content"] = line
	tags["error_code"] = fmt.Sprintf("%d", errCode)
	return metric.NewMetric("oceanbase_log", fields, tags, logAt, metric.Untyped)
}

func (e *ErrorLogInput) isFiltered(service ServiceType, line string) bool {
	// TODO: compile first
	c, found := e.config.LogServiceConfig[service]
	if found {
		for _, regex := range c.ExcludeRegexes {
			match, _ := regexp.MatchString(regex, line)
			if match {
				return true
			}
		}
	}
	return false
}

func (e *ErrorLogInput) watchFile() {
	for {
		select {
		case <-e.ctx.Done():
			log.Info("received exit signal, stop watch file routine")
			return
		default:
			// open file and set fd in file process queue
			e.watchFileChanges()
			time.Sleep(e.config.CollectDelay)
		}
	}
}

func (e *ErrorLogInput) checkAndOpenFile(logFileRealPath string) (*os.File, error) {
	var fileDesc *os.File
	_, err := os.Stat(logFileRealPath)
	if err == nil {
		fileDesc, err = os.OpenFile(logFileRealPath, os.O_RDONLY, os.ModePerm)
	}
	return fileDesc, err
}

func (e *ErrorLogInput) watchFileChanges() {
	for service, logCollectConfig := range e.config.LogServiceConfig {
		log.Infof("chekc log file of service: %s", service)
		queue, exists := e.logProcessQueue[service]
		logFileRealPath := fmt.Sprintf("%s/%s", logCollectConfig.LogConfig.LogDir, logCollectConfig.LogConfig.LogFileName)
		if exists {
			info := queue.getHead()
			if logCollectConfig.LogConfig.LogFileName == info.fileDesc.Name() {
				log.Debugf("log file of service %s not change", service)
			} else {
				log.Infof("log file of service %s has changed", service)
				// open new file and append to queue
				newFileDesc, err := e.checkAndOpenFile(logFileRealPath)
				if err != nil {
					queue.pushBack(&logFileInfo{
						fileDesc:  newFileDesc,
						isRenamed: false,
					})
				} else {
					log.Errorf("open file of service %s %s failed", service, logFileRealPath)
				}
				queue.setHeadIsRenameTrue()
			}
		} else {
			// first time, create queue, open last file
			newFileDesc, err := e.checkAndOpenFile(logFileRealPath)
			if err != nil {
				log.WithError(err).Errorf("open log file of service %s %s failed", service, logFileRealPath)
				continue
			} else {
				// initialize process queue
				q := &processQueue{
					queue: make([]*logFileInfo, 0, 8),
					mutex: sync.Mutex{},
				}
				// add new file into queue
				q.pushBack(&logFileInfo{
					fileDesc:  newFileDesc,
					isRenamed: false,
				})
				e.logProcessQueue[service] = q
			}
		}
	}
}

func (e *ErrorLogInput) Close() error {
	e.cancel()
	return nil
}

func (e *ErrorLogInput) Collect() ([]metric.Metric, error) {
	moreMetrics := true
	metrics := make([]metric.Metric, 0, 1024)
	for moreMetrics {
		select {
		case metricsFromBuffer := <-e.metricBufferChan:
			metrics = append(metrics, metricsFromBuffer...)
		default:
			log.Infof("no more metric from buffer")
			moreMetrics = false
		}
	}
	return metrics, nil
}
