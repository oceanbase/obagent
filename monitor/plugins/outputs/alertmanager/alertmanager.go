package alertmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/monitor/message"
)

const alertmanagerOutputSampleConfig = `
address: http://127.0.0.1:9093
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
	config                  *AlertmanagerOutputConfig
	httpClient              *http.Client
	taskChan                chan []*message.Message
	ctx                     context.Context
	cancelFunc              context.CancelFunc
	backgroundTaskWaitGroup sync.WaitGroup
}

func (a *AlertmanagerOutput) Init(ctx context.Context, config map[string]interface{}) error {
	rand.Seed(time.Now().UnixNano())
	configData, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "alertmanager output encode config")
	}
	a.config = &AlertmanagerOutputConfig{}
	err = yaml.Unmarshal(configData, a.config)
	if err != nil {
		return errors.Wrap(err, "alertmanager output decode config")
	}

	a.taskChan = make(chan []*message.Message, 1024)
	a.ctx, a.cancelFunc = context.WithCancel(context.Background())
	a.httpClient = &http.Client{}
	if a.config.HttpTimeout == 0 {
		a.httpClient.Timeout = defaultTimeout
	} else {
		a.httpClient.Timeout = a.config.HttpTimeout
	}

	a.backgroundTaskWaitGroup.Add(1)
	go a.schedule()

	log.Infof("alertmanager output inited with config : %v", a.config)
	return nil
}

func (a *AlertmanagerOutput) Close() error {
	a.cancelFunc()
	close(a.taskChan)
	a.backgroundTaskWaitGroup.Wait()
	return nil
}

func (a *AlertmanagerOutput) SampleConfig() string {
	return alertmanagerOutputSampleConfig
}

func (a *AlertmanagerOutput) Description() string {
	return alertmanagerOutputDescription
}

func (a *AlertmanagerOutput) Start(in <-chan []*message.Message) error {
	for messages := range in {
		a.Write(context.Background(), messages)
	}
	return nil
}

func (a *AlertmanagerOutput) Write(ctx context.Context, metrics []*message.Message) error {
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
	a.backgroundTaskWaitGroup.Done()
	for {
		select {
		case <-a.ctx.Done():
			break

		case metrics := <-a.taskChan:
			var err error
			for i := 0; i < a.config.RetryTimes; i++ {
				err = a.sendAlarm(metrics)
				if err != nil {
					log.WithError(err).Warnf("send alarm got error: %v", err)
					time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
				} else {
					break
				}
			}
		}
	}
}

func (a *AlertmanagerOutput) sendAlarm(metrics []*message.Message) error {
	alarmList := make([]map[string]interface{}, 0, a.config.BatchCount)
	for _, metricEntry := range metrics {
		alarmList = append(alarmList, a.convertMetricToAlarm(metricEntry))
	}

	jsonData, err := json.Marshal(alarmList)

	log.Debugf("send alarm metrics: %s", jsonData)

	body := bytes.NewBuffer(jsonData)
	pushAlertsAddress := fmt.Sprintf("%s/%s", a.config.Address, "api/v2/alerts")
	req, err := http.NewRequest(http.MethodPost, pushAlertsAddress, body)

	if err != nil {
		return errors.Wrap(err, "generate http request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	log.Infof("send alarm got response: %v", resp)
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

func (a *AlertmanagerOutput) convertMetricToAlarm(metric *message.Message) map[string]interface{} {
	alarmItem := make(map[string]interface{})

	labels := metric.Tags()
	labels = append(labels, message.TagEntry{"alertname", metric.GetName()})
	//labels["alertname"] = metric.GetName()
	annotations := metric.Fields()

	alarmItem["labels"] = labels
	alarmItem["annotations"] = annotations
	alarmItem["startAt"] = metric.GetTime()
	alarmItem["generatorURL"] = ""

	return alarmItem
}
