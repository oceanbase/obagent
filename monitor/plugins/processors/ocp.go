package processors

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/processors/filter"
	"github.com/oceanbase/obagent/monitor/plugins/processors/label"
)

func init() {

	plugins.GetProcessorManager().Register("excludeFilterProcessor", func(conf *monagent.PluginConfig) (plugins.Processor, error) {
		config := conf.PluginInnerConfig
		excludeFilterProcessor := &filter.ExcludeFilter{}
		err := excludeFilterProcessor.Init(context.Background(), config)
		if err != nil {
			log.WithError(err).Error("init excludeFileterProcessor failed")
			return nil, err
		}
		return excludeFilterProcessor, nil
	})
	plugins.GetProcessorManager().Register("mountLabelProcessor", func(conf *monagent.PluginConfig) (plugins.Processor, error) {
		config := conf.PluginInnerConfig
		mountLabelProcessor := &label.MountLabelProcessor{}
		err := mountLabelProcessor.Init(context.Background(), config)
		if err != nil {
			log.WithError(err).Error("Init mountLabelProcessor failed")
			return nil, err
		}
		return mountLabelProcessor, nil
	})
}
