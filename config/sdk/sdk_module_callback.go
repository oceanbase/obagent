package sdk

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	"github.com/oceanbase/obagent/config"
	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/stat"
)

// registerCallbacks To load the business callback module,
// the business needs to provide a callback function,
// and the configuration can be obtained when the business is initialized or changed.
func registerCallbacks(ctx context.Context) error {
	// register config module callbacks
	err := config.RegisterConfigCallback(
		config.NotifyProcessConfigModuleType,
		func() interface{} {
			return []config.ProcessConfigNotifyAddress{}
		},
		// Initial configuration callback
		processNotifyModuleConfigCallback,
		// Configuration update callback
		processNotifyModuleConfigCallback,
	)
	if err != nil {
		return err
	}

	err = config.RegisterConfigCallback(
		config.ConfigMetaModuleType,
		func() interface{} {
			return config.ConfigMetaBackup{}
		},
		configNotifyModuleConfigCallback,
		configNotifyModuleConfigCallback,
	)
	if err != nil {
		return err
	}

	if err := config.RegisterConfigCallback(
		config.StatConfigModuleType,
		func() interface{} {
			return stat.StatConfig{}
		},
		setStatConfig,
		setStatConfig,
	); err != nil {
		return err
	}

	return nil
}

func configNotifyModuleConfigCallback(ctx context.Context, moduleConf interface{}) error {
	conf, ok := moduleConf.(config.ConfigMetaBackup)
	if !ok {
		return errors.Errorf("module %s conf %s is not NotifyAddress", config.NotifyProcessConfigModuleType, reflect.TypeOf(moduleConf))
	}
	return config.SetConfigMetaModuleConfigNotify(ctx, conf)
}

func processNotifyModuleConfigCallback(ctx context.Context, moduleConf interface{}) error {
	confs, ok := moduleConf.([]config.ProcessConfigNotifyAddress)
	if !ok {
		return errors.Errorf("module %s conf %s is not NotifyAddress", config.NotifyProcessConfigModuleType, reflect.TypeOf(moduleConf))
	}
	for _, conf := range confs {
		config.SetProcessModuleConfigNotifyAddress(conf)
	}
	return nil
}

func setLogger(ctx context.Context, moduleConf interface{}) error {
	logconf, ok := moduleConf.(agentlog.LoggerConfig)
	if !ok {
		return errors.Errorf("conf is not agentlog.LoggerConfig")
	}
	agentlog.InitLogger(logconf)
	return nil
}

func setStatConfig(ctx context.Context, moduleConf interface{}) error {
	conf, ok := moduleConf.(stat.StatConfig)
	if !ok {
		return errors.Errorf("conf is not stat.StatConfig")
	}
	stat.SetStatConfig(ctx, conf)
	return nil
}
