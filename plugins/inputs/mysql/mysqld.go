// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package mysql

import (
	"context"

	log2 "github.com/go-kit/kit/log/logrus"
	kitLog "github.com/go-kit/log"
	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"

	"github.com/prometheus/mysqld_exporter/collector"

	"github.com/oceanbase/obagent/metric"
)

const mysqldSampleConfig = `
`

const mysqldDescription = `
`

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
	Dsn          string          `yaml:"dsn"`
	ScraperFlags map[string]bool `yaml:"scraperFlags"`
}

type MysqldExporter struct {
	Config          *MysqldConfig
	logger          kitLog.Logger
	registry        *prometheus.Registry
	collector       *collector.Exporter
	enabledScrapers []collector.Scraper
}

func (m *MysqldExporter) Close() error {
	m.registry.Unregister(m.collector)
	return nil
}

func (m *MysqldExporter) SampleConfig() string {
	return mysqldSampleConfig
}

func (m *MysqldExporter) Description() string {
	return mysqldDescription
}

func (m *MysqldExporter) Init(config map[string]interface{}) error {
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

	m.enabledScrapers = make([]collector.Scraper, 0, len(scrapers))

	for scraper, enabledByDefault := range scrapers {
		enabled, found := m.Config.ScraperFlags[scraper.Name()]
		if (found && enabled) || (!found && enabledByDefault) {
			m.enabledScrapers = append(m.enabledScrapers, scraper)
		}
	}

	ctx := context.Background()
	m.collector = collector.New(ctx, m.Config.Dsn, collector.NewMetrics(), m.enabledScrapers, m.logger)
	m.registry = prometheus.NewRegistry()
	err = m.registry.Register(m.collector)
	if err != nil {
		return errors.Wrap(err, "mysqld exporter register collector")
	}

	return err
}

func (m *MysqldExporter) Collect() ([]metric.Metric, error) {
	// TODO parse metric families

	var metrics []metric.Metric

	metricFamilies, err := m.registry.Gather()
	if err != nil {
		return nil, errors.Wrap(err, "node exporter registry gather")
	}
	for _, metricFamily := range metricFamilies {
		metricsFromMetricFamily := metric.ParseFromMetricFamily(metricFamily)
		metrics = append(metrics, metricsFromMetricFamily...)
	}

	return metrics, nil
}
