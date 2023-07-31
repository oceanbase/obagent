package prometheus

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/monitor/engine"
	"github.com/oceanbase/obagent/monitor/message"
)

const sampleConfig = `
formatType: fmtText
`

const description = `
export data using prometheus protocol
`

type Config struct {
	FormatType    string `yaml:"formatType"`
	ExposeUrl     string `yaml:"exposeUrl"`
	WithTimestamp bool   `yaml:"withTimestamp"`
}

type Prometheus struct {
	sourceConfig map[string]interface{}

	config *Config
	cache  map[string]*message.Message
}

type MetricFamily struct {
	Samples  []*Sample
	Type     message.Type
	LabelSet map[string]int
}

type Sample struct {
	Labels         map[string]string
	Value          float64
	HistogramValue map[float64]uint64
	SummaryValue   map[float64]float64
	Count          uint64
	Sum            float64
	Timestamp      time.Time
}

var Format = map[string]expfmt.Format{
	"fmtText": expfmt.FmtText,
}

func (p *Prometheus) Init(ctx context.Context, config map[string]interface{}) error {
	p.sourceConfig = config
	configData, err := yaml.Marshal(p.sourceConfig)
	if err != nil {
		return errors.Wrap(err, "prometheus exporter encode config")
	}
	p.config = &Config{}
	err = yaml.Unmarshal(configData, p.config)
	if err != nil {
		return errors.Wrap(err, "prometheus exporter decode config")
	}
	log.WithContext(ctx).Infof("prometheus exporter config: %+v", p.config)
	_, exist := Format[p.config.FormatType]
	if !exist {
		return errors.New("format type not exist")
	}
	p.cache = make(map[string]*message.Message)
	if p.config.ExposeUrl != "" {
		engine.GetRouteManager().RegisterHTTPRoute(ctx, p.config.ExposeUrl, &httpRoute{routePath: p.config.ExposeUrl})
		engine.GetRouteManager().AddPipelineGroup(p.config.ExposeUrl, p.cache)
	}
	return nil
}

func (p *Prometheus) SampleConfig() string {
	return sampleConfig
}

func (p *Prometheus) Description() string {
	return description
}

func (p *Prometheus) Start(in <-chan []*message.Message) error {
	for messages := range in {
		tmpMap := make(map[string]*message.Message)
		for _, msg := range messages {
			name := msg.Identifier()
			tmpMap[name] = msg
		}

		for k, v := range tmpMap {
			p.cache[k] = v
		}
	}
	return nil
}

func (p *Prometheus) Stop() {
	if p.config.ExposeUrl != "" {
		log.Infof("route %s cache deleting", p.config.ExposeUrl)
		err := engine.GetRouteManager().DeletePipelineGroup(p.config.ExposeUrl, p.cache)
		if err != nil {
			log.Errorf("delete route %s cache failed, err:%+v", p.config.ExposeUrl, err)
			return
		}
		log.Infof("route %s cache deleted", p.config.ExposeUrl)
	}
}
