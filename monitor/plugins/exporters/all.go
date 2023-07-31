package exporters

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/exporters/prometheus"
)

func init() {
	plugins.GetExporterManager().Register("prometheusExporter", func(conf *monagent.PluginConfig) (plugins.Sink, error) {
		prometheusExporter := &prometheus.Prometheus{}
		err := prometheusExporter.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("init prometheusExporter failed")
			return nil, err
		}
		return prometheusExporter, nil
	})
}
