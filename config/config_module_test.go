// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/lib/crypto"
	agentlog "github.com/oceanbase/obagent/log"
)

var (
	fooServer = &FooService{
		Foo: &Foo{Foo: "foo_value"},
	}
)

const (
	testFooModule                = "foo"
	testFooModuleType ModuleType = "fooType"
)

type Foo struct {
	Foo string `json:"foo"`
	Bar Bar    `json:"bar"`
}

type Bar struct {
	Bar      int    `json:"bar"`
	Duration string `json:"duration"`
}

type FooService struct {
	Foo         *Foo
	FooNoDefine string `json:"fooNoDefine"`
}

var (
	fooKVYaml = `configVersion: "foo"
configs:
    - key: foo.foo
      value: foo_value
    - key: foo.no.defined
      valueType: string
      value: foo-no-define
    - key: foo.bar.duration
      value: 10s
    - key: foo.bar.bar
      value: 3306`
)

func _init(t *testing.T) string {
	SetConfigPropertyMeta(
		&ConfigProperty{
			Key:          "foo.foo",
			DefaultValue: "foo",
			ValueType:    ValueString,
			Encrypted:    true,
			Fatal:        false,
			Masked:       false,
			NeedRestart:  false,
			Description:  "",
			Unit:         "",
			Valid:        nil,
		})
	SetConfigPropertyMeta(
		&ConfigProperty{
			Key:          "foo.bar.bar",
			DefaultValue: 3306,
			ValueType:    ValueInt64,
			Fatal:        false,
			Masked:       true,
			NeedRestart:  true,
			Description:  "",
			Unit:         "",
			Valid:        nil,
		})
	SetConfigPropertyMeta(
		&ConfigProperty{
			Key:          "foo.bar.duration",
			DefaultValue: "100ms",
			ValueType:    ValueString,
			Fatal:        false,
			Masked:       true,
			NeedRestart:  true,
			Description:  "",
			Unit:         "",
			Valid:        nil,
		})

	croptoErr := InitCrypto("", crypto.PLAIN)
	assert.Nil(t, croptoErr)

	tempDir := os.TempDir()

	moduleConfigDir := filepath.Join(tempDir, "module_config")
	err := os.MkdirAll(moduleConfigDir, 0755)
	assert.Nil(t, err)

	configPropertiesDir := filepath.Join(tempDir, "config_properties")
	err = os.MkdirAll(configPropertiesDir, 0755)
	assert.Nil(t, err)
	err = ioutil.WriteFile(filepath.Join(configPropertiesDir, "foo.yaml"), []byte(fooKVYaml), 0755)
	assert.Nil(t, err)

	err = InitModuleConfigs(moduleConfigDir)
	assert.Nil(t, err)
	err = InitConfigProperties(configPropertiesDir)
	assert.Nil(t, err)

	agentlog.InitLogger(agentlog.LoggerConfig{
		Output: os.Stdout,
		Level:  "debug",
	})

	registerFooCallback()

	return tempDir
}

func cleanup() {
	mainModuleConfig = nil
	mainConfigProperties = nil
	configPropertyMetas = map[string]*ConfigProperty{}
	callbacks = map[ModuleType]*ConfigCallback{}
}

func registerFooCallback() {

	RegisterConfigCallback(testFooModuleType,
		// create an instance of foo
		func() interface{} {
			return new(Foo)
		},
		// init config
		func(ctx context.Context, conf interface{}) error {
			log.WithField("module config", conf).Info("init foo config")
			foo, ok := conf.(*Foo)
			if !ok {
				return errors.Errorf("config is not *Foo, but %s", reflect.TypeOf(conf))
			}
			log.WithField("module", foo).Info("init foo config")
			fooServer.Foo = foo
			log.Infof("init module %s config sucessfully", testFooModule)
			return nil
		},
		// notify updated config
		func(ctx context.Context, conf interface{}) error {
			log.WithField("module config", conf).Info("update foo config")
			foo, ok := conf.(*Foo)
			if !ok {
				return errors.Errorf("config is not *Foo, but %s", reflect.TypeOf(conf))
			}
			log.WithField("module", foo).Info("update foo config")
			fooServer.Foo = foo
			log.Infof("update module %s config sucessfully", testFooModule)
			return nil
		},
	)
}

func TestModuleConfigCallback_Success(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("foo callback", func(t *testing.T) {
		callback, ex := getModuleCallback(testFooModule)
		Convey("getModuleCallback", t, func() {
			So(ex, ShouldBeTrue)
			So(callback, ShouldNotBeNil)
		})

		err := callback.InitConfigCallback(context.Background(), testFooModule)
		Convey("InitConfigCallback before UpdateConfigCallback", t, func() {
			So(err, ShouldBeNil)
			So(fooServer.Foo.Foo, ShouldEqual, "foo_value")
			So(fooServer.Foo.Bar.Bar, ShouldEqual, 3306)
			So(fooServer.Foo.Bar.Duration, ShouldEqual, "10s")
		})

	})
}

func TestGetFinalModuleConfig(t *testing.T) {
	_init(t)
	defer cleanup()

	Convey("GetFinalModuleConfig with exist module", t, func() {
		conf, err := GetFinalModuleConfig("foo")
		So(err, ShouldBeNil)
		foo, ok := conf.Config.(*Foo)
		So(ok, ShouldBeTrue)
		So(foo.Foo, ShouldEqual, "foo_value")
		So(foo.Bar.Bar, ShouldEqual, 3306)
		So(foo.Bar.Duration, ShouldEqual, "10s")
	})

	Convey("GetFinalModuleConfig with not exist module", t, func() {
		_, err := GetFinalModuleConfig("module-not-exist")
		So(err, ShouldNotBeNil)
	})

	Convey("GetFinalModuleConfig with not exist module callback", t, func() {
		delete(callbacks, "foo")
		_, err := GetFinalModuleConfig("foo")
		So(err, ShouldBeNil)
	})

}

func TestInitModule(t *testing.T) {
	_init(t)
	defer cleanup()

	err := InitModuleTypeConfig(context.Background(), testFooModuleType)
	Convey("InitModuleTypeConfig", t, func() {
		So(err, ShouldBeNil)
	})

	err = InitModuleConfig(context.Background(), testFooModule)
	Convey("InitModuleConfig", t, func() {
		So(err, ShouldBeNil)
	})
}
