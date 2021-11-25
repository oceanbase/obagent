package prometheus

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"

	// "github.com/avast/retry-go/v3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/metric"
)

const alertmanagerOutputSampleConfig = `
address: http://1.1.1.1:9093
httpTimeout: 10s
batchCount: 100
retryTimes: 3
`

const alertmanagerOutputDescription = `
send metric data as alarm to alertmanager
`

var defaultTimeout = 10 * time.Second

type AlertmanagerOutputConfig struct {
	Address     string        `yaml:"address"`
	BatchCount  int           `yaml:"batchCount"`
	HttpTimeout time.Duration `yaml:"httpTimeout"`
	RetryTimes  int           `yaml:"retryTimes"`
}

type AlertmanagerOutput struct {
	config     *AlertmanagerOutputConfig
	httpClient *http.Client
	taskChan   chan []metric.Metric
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (a *AlertmanagerOutput) Init(config map[string]interface{}) error {
	configData, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "alertmanager output encode config")
	}
	a.config = &AlertmanagerOutputConfig{}
	err = yaml.Unmarshal(configData, a.config)
	if err != nil {
		return errors.Wrap(err, "alertmanager output decode config")
	}

	a.taskChan = make(chan []metric.Metric, 1024)
	a.ctx, a.cancelFunc = context.WithCancel(context.Background())
	a.httpClient = &http.Client{}
	if a.config.HttpTimeout == 0 {
		a.httpClient.Timeout = defaultTimeout
	} else {
		a.httpClient.Timeout = a.config.HttpTimeout
	}

	go a.schedule()

	log.Infof("alertmanager output inited with config : %v", a.config)
	return nil
}

func (a *AlertmanagerOutput) Close() error {
	a.cancelFunc()
	close(a.taskChan)
	return nil
}

func (a *AlertmanagerOutput) SampleConfig() string {
	return alertmanagerOutputSampleConfig
}

func (a *AlertmanagerOutput) Description() string {
	return alertmanagerOutputDescription
}

func (a *AlertmanagerOutput) Write(metrics []metric.Metric) error {
	for len(metrics) > 0 {
		count := a.config.BatchCount
		if len(metrics) < count {
			count = len(metrics)
		}
		a.taskChan <- metrics[0:count]
		metrics = metrics[count:]
	}
	return nil
}

func (a *AlertmanagerOutput) schedule() {
	for {
		select {
		case <-a.ctx.Done():
			break

		case metrics := <-a.taskChan:
			a.sendAlarm(metrics)
		}
	}
}

func (a *AlertmanagerOutput) sendAlarm(metrics []metric.Metric) error {

	alarmList := make([]map[string]interface{}, 0, a.config.BatchCount)
	for _, metricEntry := range metrics {
		alarmList = append(alarmList, a.convertMetricToAlarm(metricEntry))
	}

	jsonData, err := json.Marshal(alarmList)
	body := bytes.NewBuffer(jsonData)
	req, err := http.NewRequest(http.MethodPost, a.config.Address, body)

	if err != nil {
		return errors.Wrap(err, "generate http request")
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "do query")
	}

	if err != nil {
		return errors.Wrap(err, "read response")
	}

	if resp.StatusCode != 200 {
		return errors.New("send alarm got abnormal status code")
	}

	return nil
}

func (a *AlertmanagerOutput) convertMetricToAlarm(metric metric.Metric) map[string]interface{} {
	alarmItem := make(map[string]interface{})

	labels := metric.Tags()
	labels["alertname"] = metric.GetName()
	annotations := metric.Fields()

	alarmItem["labels"] = labels
	alarmItem["annotations"] = annotations
	alarmItem["startAt"] = metric.GetTime()
	alarmItem["generatorURL"] = ""

	return alarmItem
}
