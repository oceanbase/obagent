package outputs

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/outputs/es"
	"github.com/oceanbase/obagent/monitor/plugins/outputs/pushhttp"
)

func init() {
	plugins.GetOutputManager().Register("httpOutput", func(conf *monagent.PluginConfig) (plugins.Sink, error) {
		httpOutput := &pushhttp.HttpOutput{}
		err := httpOutput.Init(context.Background(), conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("init httpOutput failed")
			return nil, err
		}
		return httpOutput, nil
	})
	plugins.GetOutputManager().Register("esOutput", func(conf *monagent.PluginConfig) (plugins.Sink, error) {
		esOutput, err := es.NewESOutput(conf.PluginInnerConfig)
		if err != nil {
			log.WithError(err).Error("NewESOutput failed")
			return nil, err
		}
		return esOutput, nil
	})

}
