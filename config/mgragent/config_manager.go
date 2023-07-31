package mgragent

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/lib/file"
)

type Manager struct {
	config ManagerConfig

	//propertyDefinitions map[string][]*ConfigProperty
	//
	//propertyValues      map[string][]interface{}
	//lock                sync.RWMutex
	//configGroups        []*ConfigPropertiesGroup
	//allConfigProperties map[string]*ConfigProperty
	//needRestartModules  map[string]*RestartModuleKeyValues
	//moduleConfigGroups  []*ModuleConfigGroup
	//allModuleConfigs    map[string]ModuleConfig
}

type ManagerConfig struct {
	ModuleConfigDir     string
	ConfigPropertiesDir string
	//CryptoPath          string
	//CryptoMethod        string
}

var GlobalConfigManager *Manager
var fileImpl = &file.FileImpl{}

func NewManager(config ManagerConfig) *Manager {
	return &Manager{
		config: config,
	}
}

type ModuleConfigUpdateOp string

const (
	ModuleConfigSet     ModuleConfigUpdateOp = "set"
	ModuleConfigDel     ModuleConfigUpdateOp = "delete"
	ModuleConfigRestore ModuleConfigUpdateOp = "restore"
)

type ModuleConfigChange struct {
	Operation ModuleConfigUpdateOp `json:"operation"`
	FileName  string               `json:"fileName"`
	Content   string               `json:"content"`
}

type ModuleConfigChangeRequest struct {
	ModuleConfigChanges []ModuleConfigChange `json:"moduleConfigChanges"`
	Reload              bool                 `json:"reload"`
}

// ChangeModuleConfigs change module_config files
// req contains multiple change operation.
// set means write a file to module_config dir.
// del means delete a file from module_config dir.
func (m *Manager) ChangeModuleConfigs(ctx context.Context, req *ModuleConfigChangeRequest) ([]string, error) {
	var changedFileNames []string
	if req == nil || len(req.ModuleConfigChanges) == 0 {
		log.WithContext(ctx).Info("no module configs to changes")
		return changedFileNames, nil
	}
	log.WithContext(ctx).Infof("%d files to change", len(req.ModuleConfigChanges))
	fileMap := make(map[string]bool)
	for _, change := range req.ModuleConfigChanges {
		if _, ok := fileMap[change.FileName]; ok {
			return nil, fmt.Errorf("fileName '%s' in changes duplicated", change.FileName)
		}
		fileMap[change.FileName] = true
	}

	var toApply []ModuleConfigChange
	for _, change := range req.ModuleConfigChanges {
		if change.Operation == ModuleConfigSet {
			err := m.validateYaml(change.Content)
			if err != nil {
				return nil, fmt.Errorf("validate new content of '%s' failed: %v", change.FileName, err)
			}
			changed, err := m.changed(change.FileName, change.Content)
			if err != nil {
				return nil, fmt.Errorf("check content changed of '%s' failed: %v", change.FileName, err)
			}
			if !changed {
				log.WithContext(ctx).Infof("content of '%s' not changed, nothing to do", change.FileName)
				continue
			}
			err = m.backupFile(ctx, change)
			if err != nil {
				return nil, err
			}
		} else if change.Operation == ModuleConfigDel {
			origExists, err := m.originExists(change.FileName)
			if err != nil {
				return nil, fmt.Errorf("check origin file '%s' exists failed: %v", change.FileName, err)
			}
			if !origExists {
				log.WithContext(ctx).Infof("file '%s' not exists, nothing to do", change.FileName)
				continue
			}
			err = m.backupFile(ctx, change)
			if err != nil {
				return nil, err
			}
		} else if change.Operation == ModuleConfigRestore {
			backupExists, err := m.backupExists(change.FileName)
			if err != nil {
				return nil, fmt.Errorf("check backup file '%s' exists failed %v", change.FileName, err)
			}
			if !backupExists {
				log.WithContext(ctx).Infof("backup file for '%s' not exists, nothing to do", change.FileName)
				continue
			}
		} else {
			return nil, fmt.Errorf("invalid change operation: %s", change.Operation)
		}
		toApply = append(toApply, change)
	}
	for _, change := range toApply {
		filePath := m.moduleConfigFilePath(change.FileName)
		var err error
		if change.Operation == ModuleConfigSet {
			log.WithContext(ctx).Infof("setting module config file %s", change.FileName)
			err = ioutil.WriteFile(filePath, []byte(change.Content), 0644)
		} else if change.Operation == ModuleConfigDel {
			log.WithContext(ctx).Infof("deleting module config file %s", change.FileName)
			err = m.deleteFile(ctx, change.FileName)
		} else if change.Operation == ModuleConfigRestore {
			err = m.restoreBackupFile(ctx, change.FileName)
		}
		if err != nil {
			return changedFileNames, err
		}
		changedFileNames = append(changedFileNames, change.FileName)
	}
	return changedFileNames, nil
}

func (m *Manager) changed(fileName string, newContent string) (bool, error) {
	origFile := m.moduleConfigFilePath(fileName)
	bytesContent, err := ioutil.ReadFile(origFile)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return string(bytesContent) != newContent, nil
}

func (m *Manager) deleteFile(ctx context.Context, fileName string) error {
	origFile := m.moduleConfigFilePath(fileName)
	err := os.Remove(origFile)
	if err != nil && os.IsNotExist(err) {
		// Allows to delete nonexistent files
		log.WithContext(ctx).Warnf("delete module config file '%s' got file not exists: %v", fileName, err)
		return nil
	}
	return err
}

func (m *Manager) backupFile(ctx context.Context, change ModuleConfigChange) error {
	origFile := m.moduleConfigFilePath(change.FileName)
	backupFile := m.moduleConfigBackupFilePath(change.FileName)
	backupExists, err := m.backupExists(change.FileName)
	if err != nil {
		return fmt.Errorf("check backup file %s exists failed %v", backupFile, err)
	}
	origExists, err := m.originExists(change.FileName)
	if err != nil {
		return fmt.Errorf("check origin file %s exists failed %v", origFile, err)
	}
	if !backupExists {
		if origExists {
			log.WithContext(ctx).Infof("backup module config file %s", change.FileName)
			err = fileImpl.CopyFile(origFile, backupFile, 0644)
		} else {
			log.WithContext(ctx).Infof("create empty backup module config file %s", change.FileName)
			err = ioutil.WriteFile(backupFile, []byte{}, 0644)

		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) backupExists(fileName string) (bool, error) {
	backupFile := m.moduleConfigBackupFilePath(fileName)
	return fileImpl.FileExists(backupFile)
}

func (m *Manager) originExists(fileName string) (bool, error) {
	origFile := m.moduleConfigFilePath(fileName)
	return fileImpl.FileExists(origFile)
}

func (m *Manager) restoreBackupFile(ctx context.Context, fileName string) error {
	backupFile := m.moduleConfigBackupFilePath(fileName)
	targetFile := m.moduleConfigFilePath(fileName)
	content, err := ioutil.ReadFile(backupFile)
	if err != nil {
		return err
	}
	if len(content) == 0 {
		// An empty backup file indicates that the configuration file does not exist. restore should delete it
		err = os.Remove(targetFile)
		if err == nil {
			err = os.Remove(backupFile)
		}
	} else {
		log.WithContext(ctx).Infof("restoring module config file %s", fileName)
		err = os.Rename(backupFile, targetFile)
	}
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) ReloadModuleConfigs(ctx context.Context) error {
	return config.InitModuleConfigs(ctx, m.config.ModuleConfigDir)
}

func (m *Manager) moduleConfigFilePath(fileName string) string {
	return filepath.Join(m.config.ModuleConfigDir, fileName)
}

func (m *Manager) moduleConfigBackupFilePath(fileName string) string {
	return m.moduleConfigFilePath(fileName) + ".bak"
}

func (m *Manager) validateYaml(content string) error {
	var t = &yaml.Node{}
	return yaml.Unmarshal([]byte(content), &t)
}
