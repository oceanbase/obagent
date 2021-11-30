package log

import (

	"os"
    "fmt"
    "sync"
    "bufio"
	"regexp"
    "context"
	"time"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/oceanbase/obagent/metric"

)

const sampleConfig = `
maxDelay: 300s
batchCount: 1000
`

const description = `
collect ob error logs and filter by keywords
`

type ServiceType string

const (
    RootService ServiceType="rootservice"
    Observer ServiceType="observer"
    Election ServiceType="election"
)

type LogCollectConfig struct {
    LogConfig *LogConfig `yaml:"logConfig"`
    KeywordsExclude []string `yaml:"keywordsExclude"`
}

type Config struct {
    LogServiceConfig map[ServiceType]*LogCollectConfig `yaml:"logServiceConfig"`
    CollectDelay time.Duration `yaml:"collectDelay"`
	ExpireTime   time.Duration      `yaml:"expireTime"`
	BatchCount int `yaml:"batchCount"`
}

type ErrorLogInput struct {
	config      *Config
	logAnalyzer ILogAnalyzer
    logProcessQueue map[ServiceType]*processQueue
    ctx context.Context
    cancel context.CancelFunc
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

    for service, _ := range(e.config.LogServiceConfig) {
        go e.doCollect(service)
    }

	log.Info("error log input init with config", e.config)

    return nil
}

func (e *ErrorLogInput) doCollect(service ServiceType) {
    for {
        select {
        case <- e.ctx.Done():
            log.Infof("received exit signal, stop collect routine of service %s", service)
        default:
            e.collectErrorLogs(service)
            time.Sleep(time.Second)
        }
    }
}

func (e *ErrorLogInput) collectErrorLogs(service ServiceType) {
    // TODO: read error logs and filter by keyword
    q, found := e.logProcessQueue[service]
    if !found {
        log.Warnf("service %s has no process queue", service)
    } else {
        if q.getQueueLen() == 0 {
            log.Warnf("service %s has no process queue", service)
        } else {
            // read head of queue
            logFileDesc := q.getHead()
            fdScanner := bufio.NewScanner(logFileDesc)
        }
    }
}

func (l *LogTailer) processLogByLine(ctx context.Context, fd *os.File, tailConf *config.TailConfig, reportedErrMap map[int]*reportedError) error {
        ctxLog := log.WithContext(ctx).WithFields(log.Fields{
                "fileName":       fd.Name(),
                "tailConf":       tailConf,
                "reportedErrMap": reportedErrMap,
        })
        // 每次读取的大小不能超过 64k，默认按行分割
        fdScanner := bufio.NewScanner(fd)
        for fdScanner.Scan() {
                line := fdScanner.Text()
                if line == "" || len(line) == 0 {
                        continue
                }
                err := l.processLogAlarm(ctx, line, tailConf, reportedErrMap)
                if err != nil {
                        ctxLog.WithField("line", line).WithError(err).Warn("failed to processLogAlarm")
                        continue
                }
                select {
                case _, isOpen := <-l.isStop:
                        if !isOpen {
                                ctxLog.Info("stop scan file")
                                return nil
                        }
                default:
                }
        }
        if err := fdScanner.Err(); err != nil {
                ctxLog.WithError(err).Error("failed to scan")
                return err
        }
        return nil
}

func (e *ErrorLogInput) watchFile() {
    for {
        select {
        case <- e.ctx.Done():
            log.Info("received exit signal, stop watch file routine")
            // TODO: close opened file
            return
        default:
            // open file and set fd in file process queue
            e.watchFileChanges()
            time.Sleep(time.Second)
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
                        fileDesc: newFileDesc,
                        isRenamed: false,
                    })
                } else {
                    log.Errorf("open file of service %s %s failed", service, logFileRealPath)
                }
                info.isRenamed = true
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
                    fileDesc: newFileDesc,
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
        case metricsFromBuffer := <- e.metricBufferChan:
            metrics = append(metrics, metricsFromBuffer...)
        default:
            log.Infof("no more metric from buffer")
            moreMetrics = false
        }
    }
    return metrics, nil
}

type filterRuleInfo struct {
	reg        *regexp.Regexp
	expireTime time.Time
}

