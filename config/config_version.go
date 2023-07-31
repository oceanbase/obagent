package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	configVersionFormat = "2006-01-02T15:04:05.9999Z07:00"
)

var configMetaBackupWorker *checkConfigVersionBackupWorker

func init() {
	configMetaBackupWorker = &checkConfigVersionBackupWorker{}
}

type ConfigMetaBackup struct {
	MaxBackups int64 `yaml:"maxbackups"`
}

func SetConfigMetaModuleConfigNotify(ctx context.Context, conf ConfigMetaBackup) error {
	if conf.MaxBackups < 0 {
		return errors.Errorf("maxbuckups must be bigger than 0.")
	}
	log.WithContext(ctx).Infof("receive version backup conf: %+v", conf)
	configMetaBackupWorker.configMetaBackup = conf

	atomic.StoreInt32(&configMetaBackupWorker.working, 1)

	go configMetaBackupWorker.checkOnce(ctx)
	return nil
}

type checkConfigVersionBackupWorker struct {
	working int32

	configMetaBackup ConfigMetaBackup
}

func (worker *checkConfigVersionBackupWorker) checkOnce(ctx context.Context) {
	if atomic.LoadInt32(&worker.working) <= 0 {
		log.WithContext(ctx).Infof("config version backup worker is not working in current process.")
		return
	}
	log.WithContext(ctx).Info("check config version backups once.")
	configDir := filepath.Dir(mainConfigProperties.configPropertiesDir)
	err := checkConfigVersionBackups(ctx, int(worker.configMetaBackup.MaxBackups), configDir)
	if err != nil {
		log.WithContext(ctx).Error(err)
	}
}

func generateNewConfigVersion() *ConfigVersion {
	return &ConfigVersion{
		ConfigVersion: time.Now().Format(configVersionFormat),
	}
}

func isConfigVersionDir(file os.FileInfo) (bool, time.Time) {
	if !file.IsDir() {
		return false, time.Now()
	}
	t, err := time.Parse(configVersionFormat, file.Name())
	return err == nil, t
}

type configVersionFile struct {
	fileName string
	version  time.Time
}

func checkConfigVersionBackups(ctx context.Context, versionStayCount int, configPath string) error {
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		return errors.Errorf("read config files from dir:%s, err:%s", configPath, err)
	}
	log.WithContext(ctx).Infof("read config from path %s, files length %d", configPath, len(files))
	if len(files) <= 0 {
		log.WithContext(ctx).Infof("no config file exists in path %s", configPath)
		return nil
	}
	configVersionFiles := make([]configVersionFile, 0, 20)
	for _, file := range files {
		log.WithContext(ctx).Debugf("read config from path %s, file %s", configPath, file.Name())
		if ok, tm := isConfigVersionDir(file); ok {
			configVersionFiles = append(configVersionFiles, configVersionFile{
				fileName: filepath.Join(configPath, file.Name()),
				version:  tm,
			})
		}
	}
	if len(configVersionFiles) <= versionStayCount {
		log.WithContext(ctx).Infof("config stay count:%d, config version count:%d, no need to rotate.", versionStayCount, len(configVersionFiles))
		return nil
	}

	sort.Slice(configVersionFiles, func(i, j int) bool {
		return configVersionFiles[i].version.After(configVersionFiles[j].version)
	})

	for i := versionStayCount; i < len(configVersionFiles); i++ {
		log.WithContext(ctx).Infof("remove config version file:%s", configVersionFiles[i].fileName)
		err := os.RemoveAll(configVersionFiles[i].fileName)
		if err != nil {
			log.WithContext(ctx).Errorf("remove config file %s failed, err:%+v", configVersionFiles[i].fileName, err)
		}
	}

	return nil
}
