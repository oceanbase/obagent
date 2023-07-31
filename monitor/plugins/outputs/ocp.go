package outputs

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/config/monagent"
	"github.com/oceanbase/obagent/monitor/plugins"
	"github.com/oceanbase/obagent/monitor/plugins/outputs/sls"
)

func init() {
	plugins.GetOutputManager().Register("slsOutput", func(conf *monagent.PluginConfig) (plugins.Sink, error) {
		configBytes, err := yaml.Marshal(conf.PluginInnerConfig)
		if err != nil {
			log.Errorf("marshal slsOutput config failed, err: %s", err)
			return nil, err
		}
		var slsOutputConfig = &sls.Config{}
		err = yaml.Unmarshal(configBytes, slsOutputConfig)
		if err != nil {
			log.Errorf("unmarshal slsOutput config failed, err: %s", err)
			return nil, err
		}
		slsOutput := sls.NewSLSOutput(slsOutputConfig)
		return slsOutput, nil
	})
}
