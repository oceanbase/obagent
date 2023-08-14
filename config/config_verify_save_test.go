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
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestSaveConfigSnapshot(t *testing.T) {
	tempDir := _init(t)
	defer cleanup()

	configVersion := "foo"
	snapshotPath := filepath.Join(tempDir, configVersion)

	t.Run("save config snapshot with foo (path not exist before save)", func(t *testing.T) {
		os.RemoveAll(snapshotPath)

		err := snapshotForConfigVersion(context.Background(), configVersion)
		Convey("saveConfigSnapshot", t, func() {
			So(err, ShouldBeNil)
		})

		version := GetCurrentConfigVersion()
		Convey("get current config version", t, func() {
			So(version, ShouldNotBeNil)
			So(version.ConfigVersion, ShouldEqual, configVersion)
		})

		tmpModuleConfigs, err := decodeModuleConfigGroups(context.Background(), filepath.Join(snapshotPath, "module_config"))
		Convey("DecodeModuleConfig", t, func() {
			So(err, ShouldBeNil)
			assert.NotSame(t, mainModuleConfig.allModuleConfigs, tmpModuleConfigs.allModuleConfigs)
			So(
				reflect.DeepEqual(
					mainModuleConfig.allModuleConfigs,
					tmpModuleConfigs.allModuleConfigs,
				),
				ShouldBeTrue,
			)

			_, err := GetFinalModuleConfig(testFooModule)
			So(err, ShouldBeNil)

			fooModule, ex := GetModuleConfigs()[testFooModule]
			So(ex, ShouldBeTrue)

			tmpFooModule, ex := tmpModuleConfigs.allModuleConfigs[testFooModule]
			So(ex, ShouldBeTrue)
			So(
				reflect.DeepEqual(
					fooModule,
					tmpFooModule,
				),
				ShouldBeTrue,
			)
		})
	})

	t.Run("save config snapshot with foo (path already exist before save)", func(t *testing.T) {
		err := snapshotForConfigVersion(context.Background(), configVersion)
		Convey("saveConfigSnapshot", t, func() {
			So(err, ShouldBeNil)
		})
	})

}

func TestSaveConfigProperties(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("change config", func(t *testing.T) {
		diffKVs := map[string]interface{}{
			"foo.foo":          "diff-foo",
			"foo.bar.duration": "100h",
			"key-not-exist":    "xxx",
		}
		configVersion, err := saveIncrementalConfig(context.Background(), diffKVs)
		Convey("saveIncrementalConfig", t, func() {
			So(err, ShouldBeNil)
		})
		log.Infof("configVersion %s", configVersion)

		properties := GetConfigPropertiesKeyValues()
		Convey("GetConfigPropertiesKeyValues", t, func() {
			So(properties["foo.foo"], ShouldEqual, "diff-foo")
			So(properties["foo.bar.duration"], ShouldEqual, "100h")
			_, ex := properties["key-not-exist"]
			So(ex, ShouldBeFalse)
		})
	})

	t.Run("reload config", func(t *testing.T) {
		err := ReloadConfigFromFiles(context.Background())
		Convey("ReloadConfigFromFiles", t, func() {
			So(err, ShouldBeNil)
		})

		properties := GetConfigPropertiesKeyValues()
		Convey("GetConfigPropertiesKeyValues", t, func() {
			So(properties["foo.foo"], ShouldEqual, "diff-foo")
			So(properties["foo.bar.duration"], ShouldEqual, "100h")
			_, ex := properties["key-not-exist"]
			So(ex, ShouldBeFalse)
		})
	})
}

func TestGetUpdatedConfigs(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("get updated configs", func(t *testing.T) {
		diffKVs := map[string]interface{}{
			"foo.foo":          "diff-foo",
			"foo.bar.duration": "100h",
			"key-not-exist":    "xxx",
		}
		updatedModules, err := getUpdatedConfigs(context.Background(), diffKVs)
		Convey("getUpdatedConfigs", t, func() {
			So(err, ShouldBeNil)
			So(len(updatedModules), ShouldEqual, 1)
			So(updatedModules[0].Module, ShouldEqual, testFooModule)
			So(updatedModules[0].Process, ShouldEqual, ProcessManagerAgent)
			So(reflect.DeepEqual(updatedModules[0].UpdatedKeyValues, map[string]interface{}{
				"foo.foo":          "diff-foo",
				"foo.bar.duration": "100h",
			}), ShouldBeTrue)
		})
	})
}

func TestVerifyAndSaveConfig(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("verify and save diff kvs success", func(t *testing.T) {
		diffKVs := map[string]interface{}{
			"foo.foo":          "diff-foo",
			"foo.bar.duration": "100h",
		}
		result, err := verifyAndSaveConfig(context.Background(), diffKVs)
		Convey("verifyAndSaveConfig", t, func() {
			So(err, ShouldBeNil)
			So(len(result.UpdatedConfigs), ShouldEqual, 1)
			So(result.UpdatedConfigs[0].Module, ShouldEqual, testFooModule)
			So(result.UpdatedConfigs[0].Process, ShouldEqual, ProcessManagerAgent)
			So(reflect.DeepEqual(result.UpdatedConfigs[0].UpdatedKeyValues, map[string]interface{}{
				"foo.foo":          "diff-foo",
				"foo.bar.duration": "100h",
			}), ShouldBeTrue)
		})

		properties := GetConfigPropertiesKeyValues()
		Convey("GetConfigPropertiesKeyValues", t, func() {
			So(properties["foo.foo"], ShouldEqual, "diff-foo")
			So(properties["foo.bar.duration"], ShouldEqual, "100h")
			_, ex := properties["key-not-exist"]
			So(ex, ShouldBeFalse)
		})
	})

	t.Run("verify with no exist key", func(t *testing.T) {
		diffKVs := map[string]interface{}{
			"foo.foo":          "diff-foo",
			"foo.bar.duration": "100h",
			"key-not-exist":    "xxx",
		}
		_, err := verifyAndSaveConfig(context.Background(), diffKVs)
		Convey("verifyAndSaveConfig", t, func() {
			So(err, ShouldNotBeNil)
		})
	})

	t.Run("verify with no keys", func(t *testing.T) {
		diffKVs := map[string]interface{}{}
		_, err := verifyAndSaveConfig(context.Background(), diffKVs)
		Convey("verifyAndSaveConfig", t, func() {
			So(err, ShouldNotBeNil)
		})
	})

	t.Run("verify with wrong valueType key", func(t *testing.T) {
		diffKVs := map[string]interface{}{
			"foo.foo":          100,
			"foo.bar.bar":      "should be int64",
			"foo.bar.duration": "100h",
		}
		_, err := verifyAndSaveConfig(context.Background(), diffKVs)
		Convey("verifyAndSaveConfig", t, func() {
			So(err, ShouldNotBeNil)
		})
	})
}

func TestReloadConfigFromFiles_Fail(t *testing.T) {
	tempDir := _init(t)
	defer cleanup()

	t.Run("reload config from no exist file", func(t *testing.T) {
		noExistPath := filepath.Join(tempDir, "no-exist-path")
		mainModuleConfig.moduleConfigDir = noExistPath
		err := ReloadConfigFromFiles(context.Background())
		Convey("ReloadConfigFromFiles", t, func() {
			So(err, ShouldNotBeNil)
		})

		mainConfigProperties.configPropertiesDir = noExistPath
		err = ReloadConfigFromFiles(context.Background())
		Convey("ReloadConfigFromFiles", t, func() {
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidateConfig(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("validate config success", func(t *testing.T) {
		err := ValidateConfigPairs(context.Background(), []string{"foo.foo=foo_value", "foo.bar.bar=3306"})
		Convey("ValidateConfigPairs", t, func() {
			So(err, ShouldBeNil)
		})
	})

	t.Run("validate config failed", func(t *testing.T) {
		err := ValidateConfigPairs(context.Background(), []string{"foo.foo=wrong-value"})
		Convey("ValidateConfigPairs with wrong value", t, func() {
			So(err, ShouldNotBeNil)
		})

		err2 := ValidateConfigPairs(context.Background(), []string{"key-not-exist=foo"})
		Convey("ValidateConfigPairs with wrong value", t, func() {
			So(err2, ShouldNotBeNil)
		})
	})
}

func Test_parseKeyValue(t *testing.T) {
	tests := []struct {
		name    string
		pair    string
		key     string
		value   string
		success bool
	}{
		{
			name:    "normal key-value pair",
			pair:    "key1=value1",
			key:     "key1",
			value:   "value1",
			success: true,
		},
		{
			name:    "abnormal key-value pair",
			pair:    "key1=value1=",
			key:     "key1",
			value:   "value1",
			success: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, err := parseKeyValue(tt.pair)
			if tt.success {
				Convey(tt.name, t, func() {
					So(err, ShouldBeNil)
					So(key, ShouldEqual, tt.key)
					So(value, ShouldEqual, tt.value)
				})
			} else {
				Convey(tt.name, t, func() {
					So(err, ShouldNotBeNil)
				})
			}
		})
	}
}
