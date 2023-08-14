/*
 * Copyright (c) 2023 OceanBase
 * OBAgent is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package config

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ReloadConfigFromFiles Reload all configurations: configuration items, module configuration templates
func ReloadConfigFromFiles(ctx context.Context) error {
	if mainConfigProperties == nil {
		return errors.Errorf("config properties meta is lost, can not reload config properties.")
	}
	if mainModuleConfig == nil {
		return errors.Errorf("module config meta is lost, can not reload module config.")
	}

	err := InitConfigProperties(ctx, mainConfigProperties.configPropertiesDir)
	if err != nil {
		return errors.Errorf("reload config properties from path %s, err:%s", mainConfigProperties.configPropertiesDir, err)
	}

	err = InitModuleConfigs(ctx, mainModuleConfig.moduleConfigDir)
	if err != nil {
		return errors.Errorf("reload module config from path %s, err:%s", mainModuleConfig.moduleConfigDir, err)
	}

	log.WithContext(ctx).Info("reload config success.")
	return nil
}

// InitConfigProperties Initialize configuration items and parse yaml files in the entire directory
func InitConfigProperties(ctx context.Context, configPropertiesDir string) error {

	tmpConfigMain, err := decodeConfigPropertiesGroups(ctx, configPropertiesDir)
	if err != nil {
		err = errors.Errorf("decode config properties from path %s, err:%s", configPropertiesDir, err)
		log.WithContext(ctx).Error(err)
		return err
	}
	mainConfigProperties = tmpConfigMain
	mergeConfigProperties(ctx)

	log.WithContext(ctx).Info("init config properties success.")
	return nil
}

// InitModuleConfigs Initialize the module configuration: the yaml file in the entire directory will be parsed
func InitModuleConfigs(ctx context.Context, moduleConfigDir string) error {
	tmpMainModuleConfig, err := decodeModuleConfigGroups(ctx, moduleConfigDir)
	if err != nil {
		err = errors.Errorf("decode module config from path %s, err:%s", moduleConfigDir, err)
		log.WithContext(ctx).Error(err)
		return err
	}
	mainModuleConfig = tmpMainModuleConfig

	return nil
}

// Parse the entire directory of configuration items: All yaml configuration files are parsed
func decodeConfigPropertiesGroups(ctx context.Context, configPropertiesDir string) (*ConfigPropertiesMain, error) {
	absconfigPropertiesDir, err := filepath.Abs(configPropertiesDir)
	if err != nil {
		return nil, errors.Errorf("get absolute path of %s err:%s", configPropertiesDir, err)
	}
	configMain := &ConfigPropertiesMain{
		ConfigGroups:        make([]*ConfigPropertiesGroup, 0, 4),
		allConfigProperties: make(map[string]*ConfigProperty, 128),
		needRestartModules:  map[string]*RestartModuleKeyValues{},
		configPropertiesDir: absconfigPropertiesDir,
	}
	err = loadYamlFilesFromPath(ctx, absconfigPropertiesDir, func(filename string) error {
		bs, err := ioutil.ReadFile(filename)
		if err != nil {
			return errors.Errorf("read config properties file %s, err:%s", filename, err)
		}
		group, err := decodeConfigPropertiesGroup(ctx, bs)
		if err != nil {
			return errors.Errorf("decode config properties from config file %s, err:%s", filename, err)
		}
		group.ConfigFile = filename
		if err := configMain.addConfigs(group); err != nil {
			return err
		}
		return nil
	})
	return configMain, err
}

func decodeConfigPropertiesGroup(ctx context.Context, bs []byte) (*ConfigPropertiesGroup, error) {
	configGroup := new(ConfigPropertiesGroup)
	err := Decode(bytes.NewReader(bs), configGroup)
	if err != nil {
		return nil, err
	}
	for _, property := range configGroup.Configs {
		v := property.Value
		meta, ex := configPropertyMetas[property.Key]
		// if config key is not defined in sdk, use the valueType from config file
		if !ex {
			val, err := property.Parse(v)
			if err != nil {
				return nil, errors.Errorf("parse config key %s, err:%s", property.Key, err)
			}
			log.WithContext(ctx).Infof("config key %s is not defined in sdk, use valueType %s", property.Key, property.ValueType)
			property.Value = val
			continue
		}
		if meta.ValueType != property.ValueType && property.ValueType != "" {
			log.WithContext(ctx).Warnf("config key %s valueType is defined as %s, not %s", property.Key, meta.ValueType, property.ValueType)
		}
		val, err := meta.Parse(v)
		if err != nil {
			return nil, errors.Errorf("parse config key %s, err:%s", property.Key, err)
		}
		property.Value = val
	}
	return configGroup, nil
}

func decodeModuleConfigGroups(ctx context.Context, moduleConfigDir string) (*moduleConfigMain, error) {
	absModuleConfigDir, err := filepath.Abs(moduleConfigDir)
	if err != nil {
		return nil, errors.Errorf("get absolute path of %s err:%s", moduleConfigDir, err)
	}
	mainModuleConfig := &moduleConfigMain{
		moduleConfigDir:    absModuleConfigDir,
		moduleConfigGroups: make([]*ModuleConfigGroup, 0, 4),
		allModuleConfigs:   make(map[string]ModuleConfig, 10),
	}
	err = loadYamlFilesFromPath(ctx, absModuleConfigDir, func(filename string) error {
		bs, err := ioutil.ReadFile(filename)
		if err != nil {
			return errors.Errorf("read module config file %s, err:%s", filename, err)
		}
		group, err := decodeModuleConfigGroup(ctx, bs)
		if err != nil {
			return errors.Errorf("decode module config from config file %s, err:%s", filename, err)
		}
		group.ConfigFile = filename
		for _, moduleConfigGroup := range group.Modules {
			if _, ex := mainModuleConfig.allModuleConfigs[moduleConfigGroup.Module]; ex {
				return errors.Errorf("module %s already exist.", moduleConfigGroup.Module)
			}
			if err := RegisterModule(moduleConfigGroup.Module, moduleConfigGroup.ModuleType); err != nil {
				return err
			}
			mainModuleConfig.allModuleConfigs[moduleConfigGroup.Module] = moduleConfigGroup
		}
		mainModuleConfig.moduleConfigGroups = append(mainModuleConfig.moduleConfigGroups, group)
		return nil
	})
	return mainModuleConfig, err
}

func decodeModuleConfigGroup(ctx context.Context, bs []byte) (*ModuleConfigGroup, error) {
	moduleConfigs := new(ModuleConfigGroup)
	err := Decode(bytes.NewReader(bs), moduleConfigs)
	return moduleConfigs, err
}

// snapshotPath is the target directory of the snapshot. currentPath is the current configuration directory.
// You need to create a subdirectory under snapshotPath and back up files in currentPath to the subdirectory.
// If an error occurs in the middle of the process, you should keep the error for the time being and
// return to it altogether: backup all files as much as possible.
func snapshotForPath(ctx context.Context, snapshotPath string, currentPath string) error {
	basePath := filepath.Base(currentPath)
	snapshotSubPath := filepath.Join(snapshotPath, basePath)
	err := os.MkdirAll(snapshotSubPath, 0755)
	if err != nil {
		err = errors.Errorf("snapshot path %s, create path %s, err:%s", snapshotPath, snapshotSubPath, err)
		log.WithContext(ctx).Error(err)
		return err
	}

	var errs error
	err = loadYamlFilesFromPath(ctx, currentPath, func(srcFilename string) error {
		_, filename := filepath.Split(srcFilename)
		bs, err := ioutil.ReadFile(srcFilename)
		if err != nil {
			log.WithContext(ctx).Errorf("snapshot path %s, read config file %s, err:%+v", snapshotPath, srcFilename, err)
			errs = errors.Errorf("%s ,another err:%s", errs, err)
			return nil
		}
		dstFilename := filepath.Join(snapshotSubPath, filename)
		err = ioutil.WriteFile(dstFilename, bs, 0644)
		if err != nil {
			log.WithContext(ctx).Errorf("snapshot path %s, write config file to %s to snapshot file %s, err:%+v",
				snapshotPath, srcFilename, dstFilename, err)
			errs = errors.Errorf("%s ,another err:%s", errs, err)
			return nil
		}

		// return nil to continue
		return nil
	})
	if err != nil {
		return err
	}

	return errs
}

func loadYamlFilesFromPath(ctx context.Context, configPath string, operator func(filename string) error) error {
	log.WithContext(ctx).Infof("read config files from path %s", configPath)

	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		return errors.Errorf("read config files from dir:%s, err:%s", configPath, err)
	}
	log.WithContext(ctx).Infof("read config from path %s, files length %d", configPath, len(files))
	if len(files) <= 0 {
		return errors.Errorf("no config file exists in path %s", configPath)
	}
	for _, file := range files {
		log.WithContext(ctx).Debugf("read config from path %s, file %s", configPath, file.Name())
		if !isYamlFile(file) {
			log.WithContext(ctx).Debugf("read config from path %s, file %s is not yaml, skip it.", configPath, file.Name())
			continue
		}

		filename := filepath.Join(configPath, file.Name())
		err := operator(filename)
		if err != nil {
			return errors.Errorf("decode config file %s err:%s", filename, err)
		}
	}
	return nil
}

func isYamlFile(file os.FileInfo) bool {
	return !file.IsDir() &&
		(strings.HasSuffix(file.Name(), "yaml") || strings.HasSuffix(file.Name(), "yml"))
}
