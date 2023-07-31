package sdk

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/lib/crypto"
)

var (
	configPropertiesMetaOnce sync.Once
)

type SDKConfig struct {
	ConfigPropertiesDir string
	ModuleConfigDir     string
	CryptoPath          string              `yaml:"cryptoPath"`
	CryptoMethod        crypto.CryptoMethod `yaml:"cryptoMethod"`
}

// InitSDK SDK initialization:
// 1. Load key-value meta information: only value is for users to change, and other content is put into the SDK.
// 2. Initialize the key-value configuration and the configuration template of the business module.
// 3. To load the business callback module, the business needs to provide a callback function,
// and the configuration can be obtained when the business is initialized or changed.
// 4. The callback service of the configuration module may be cross-process,
// and the cross-process API can be configured. It itself serves as part of configuration management.
func InitSDK(ctx context.Context, conf config.SDKConfig) error {
	log.WithContext(ctx).Infof("init sdk conf %+v", conf)
	// init metas
	configPropertiesMetaOnce.Do(func() {
		setConfigPropertyMeta()
	})

	if err := config.InitCrypto(conf.CryptoPath, conf.CryptoMethod); err != nil {
		log.WithContext(ctx).Error(err)
		return err
	}

	// init configs
	err := initConfigs(ctx, conf.ConfigPropertiesDir, conf.ModuleConfigDir)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return err
	}

	// init config callbacks
	if err := registerCallbacks(ctx); err != nil {
		log.WithContext(ctx).Error(err)
		return err
	}

	// set process notify address
	err = config.InitModuleConfig(context.Background(), config.NotifyProcessConfigModule)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return err
	}
	return nil
}

// initConfigs Initialize the key-value configuration and the configuration template of the business module.
func initConfigs(ctx context.Context, configPropertiesDir, moduleConfigDir string) error {
	err := config.InitConfigProperties(ctx, configPropertiesDir)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return err
	}
	err = config.InitModuleConfigs(ctx, moduleConfigDir)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return err
	}
	return nil
}
