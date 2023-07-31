package mysql

import (
	"context"
	"time"

	log2 "github.com/go-kit/kit/log/logrus"
	kitLog "github.com/go-kit/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/mysqld_exporter/collector"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/monitor/message"
)

const mysqldSampleConfig = `
`

const mysqldDescription = `
`

var defaultCollectInterval = 15 * time.Second

var (
	dsn string
)

// scrapers lists all possible collection methods and if they should be enabled by default.
var scrapers = map[collector.Scraper]bool{
	collector.ScrapeGlobalStatus{}:                        true,
	collector.ScrapeGlobalVariables{}:                     true,
	collector.ScrapeSlaveStatus{}:                         true,
	collector.ScrapeProcesslist{}:                         false,
	collector.ScrapeUser{}:                                false,
	collector.ScrapeTableSchema{}:                         false,
	collector.ScrapeInfoSchemaInnodbTablespaces{}:         false,
	collector.ScrapeInnodbMetrics{}:                       false,
	collector.ScrapeAutoIncrementColumns{}:                false,
	collector.ScrapeBinlogSize{}:                          false,
	collector.ScrapePerfTableIOWaits{}:                    false,
	collector.ScrapePerfIndexIOWaits{}:                    false,
	collector.ScrapePerfTableLockWaits{}:                  false,
	collector.ScrapePerfEventsStatements{}:                false,
	collector.ScrapePerfEventsStatementsSum{}:             false,
	collector.ScrapePerfEventsWaits{}:                     false,
	collector.ScrapePerfFileEvents{}:                      false,
	collector.ScrapePerfFileInstances{}:                   false,
	collector.ScrapePerfMemoryEvents{}:                    false,
	collector.ScrapePerfReplicationGroupMembers{}:         false,
	collector.ScrapePerfReplicationGroupMemberStats{}:     false,
	collector.ScrapePerfReplicationApplierStatsByWorker{}: false,
	collector.ScrapeUserStat{}:                            false,
	collector.ScrapeClientStat{}:                          false,
	collector.ScrapeTableStat{}:                           false,
	collector.ScrapeSchemaStat{}:                          false,
	collector.ScrapeInnodbCmp{}:                           true,
	collector.ScrapeInnodbCmpMem{}:                        true,
	collector.ScrapeQueryResponseTime{}:                   true,
	collector.ScrapeEngineTokudbStatus{}:                  false,
	collector.ScrapeEngineInnodbStatus{}:                  false,
	collector.ScrapeHeartbeat{}:                           false,
	collector.ScrapeSlaveHosts{}:                          false,
	collector.ScrapeReplicaHost{}:                         false,
}

type MysqldConfig struct {
	Dsn             string          `yaml:"dsn"`
	ScraperFlags    map[string]bool `yaml:"scraperFlags"`
	CollectInterval time.Duration   `yaml:"collect_interval"`
}

type MysqldInput struct {
	Config          *MysqldConfig
	logger          kitLog.Logger
	registry        *prometheus.Registry
	collector       *collector.Exporter
	enabledScrapers []collector.Scraper
	ctx             context.Context
	done            chan struct{}
}

func (m *MysqldInput) Close() error {
	m.registry.Unregister(m.collector)
	return nil
}

func (m *MysqldInput) SampleConfig() string {
	return mysqldSampleConfig
}

func (m *MysqldInput) Description() string {
	return mysqldDescription
}

func (m *MysqldInput) Init(ctx context.Context, config map[string]interface{}) error {
	var pluginConfig MysqldConfig

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "mysqld exporter encode config")
	}

	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "mysqld exporter decode config")
	}

	m.logger = log2.NewLogrusLogger(log.StandardLogger())

	m.Config = &pluginConfig
	log.Info("table input init with config", m.Config)
	if m.Config.CollectInterval == 0 {
		m.Config.CollectInterval = defaultCollectInterval
	}

	m.ctx = ctx
	m.done = make(chan struct{})
	m.enabledScrapers = make([]collector.Scraper, 0, len(scrapers))

	for scraper, enabledByDefault := range scrapers {
		enabled, found := m.Config.ScraperFlags[scraper.Name()]
		if (found && enabled) || (!found && enabledByDefault) {
			m.enabledScrapers = append(m.enabledScrapers, scraper)
		}
	}

	m.collector = collector.New(ctx, m.Config.Dsn, collector.NewMetrics(), m.enabledScrapers, m.logger)
	m.registry = prometheus.NewRegistry()
	err = m.registry.Register(m.collector)
	if err != nil {
		return errors.Wrap(err, "mysqld exporter register collector")
	}

	return err
}

func (m *MysqldInput) Start(out chan<- []*message.Message) error {
	log.Infof("start tableInput plugin")
	go m.update(out)
	return nil
}

func (m *MysqldInput) update(out chan<- []*message.Message) {
	ticker := time.NewTicker(m.Config.CollectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			msgs, err := m.CollectMsgs(m.ctx)
			if err != nil {
				log.WithContext(m.ctx).Warnf("mysqld collect failed, reason: %s", err)
				continue
			}
			out <- msgs
		case <-m.done:
			log.Info("mysqld exporter exited")
			return
		}
	}
}

func (m *MysqldInput) Stop() {
	if m.done != nil {
		close(m.done)
	}
}

func (m *MysqldInput) CollectMsgs(ctx context.Context) ([]*message.Message, error) {

	var metrics []*message.Message

	metricFamilies, err := m.registry.Gather()
	if err != nil {
		return nil, errors.Wrap(err, "node exporter registry gather")
	}
	for _, metricFamily := range metricFamilies {
		metricsFromMetricFamily := message.ParseFromMetricFamily(metricFamily)
		metrics = append(metrics, metricsFromMetricFamily...)
	}

	return metrics, nil
}
