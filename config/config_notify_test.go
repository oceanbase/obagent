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
	"testing"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCallNotifyModuleConfigs_Success(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("foo notify module configs", func(t *testing.T) {
		err := InitModuleConfig(context.Background(), testFooModule)
		Convey("InitModuleConfig", t, func() {
			So(err, ShouldBeNil)
			So(fooServer.Foo.Foo, ShouldEqual, "foo_value")
			So(fooServer.Foo.Bar.Bar, ShouldEqual, 3306)
		})

		err = NotifyModuleConfigs(context.Background(), &VerifyConfigResult{
			UpdatedConfigs: []*NotifyModuleConfig{
				{
					Process: ProcessManagerAgent,
					Module:  testFooModule,
					Config:  &Foo{Foo: "foofoo", Bar: Bar{Bar: 2883}},
				},
			},
		})
		Convey("NotifyModuleConfigs", t, func() {
			So(err, ShouldBeNil)
			So(fooServer.Foo.Foo, ShouldEqual, "foofoo")
			So(fooServer.Foo.Bar.Bar, ShouldEqual, 2883)
		})

		err = NotifyLocalModuleConfig(context.Background(), &NotifyModuleConfig{
			Process: ProcessManagerAgent,
			Module:  testFooModule,
			Config:  &Foo{Foo: "foofoo2", Bar: Bar{Bar: 2881}},
		})
		Convey("NotifyLocalModuleConfig", t, func() {
			So(err, ShouldBeNil)
			So(fooServer.Foo.Foo, ShouldEqual, "foofoo2")
			So(fooServer.Foo.Bar.Bar, ShouldEqual, 2881)
		})

		err = NotifyModuleConfigForHttp(context.Background(), &NotifyModuleConfig{
			Process: ProcessManagerAgent,
			Module:  testFooModule,
			Config:  &Foo{Foo: "foofoo3", Bar: Bar{Bar: 2882}},
		})
		Convey("NotifyModuleConfigForHttp", t, func() {
			So(err, ShouldBeNil)
			So(fooServer.Foo.Foo, ShouldEqual, "foo_value")
			So(fooServer.Foo.Bar.Bar, ShouldEqual, 3306)
		})
	})
}

func TestInitModuleConfig_WithNoExistModule(t *testing.T) {
	_init(t)
	defer cleanup()

	err := InitModuleConfig(context.Background(), "module_not_exist")
	log.Errorf("err:%s", err)
	Convey("InitModuleConfig with no exist module", t, func() {
		So(err, ShouldNotBeNil)
	})
}

func TestNotifyRemoteModuleConfig_Fail(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("notify remote no exist process", func(t *testing.T) {
		err := NotifyRemoteModuleConfig(context.Background(), &NotifyModuleConfig{
			Process: Process("not_exist_process"),
			Module:  testFooModule,
			Config:  &Foo{Foo: "foofoo2", Bar: Bar{Bar: 2884}},
		})
		log.Errorf("err:%s", err)
		Convey("NotifyRemoteModuleConfig with err", t, func() {
			So(err, ShouldNotBeNil)
		})
	})

	t.Run("notify remote exist process but without process config", func(t *testing.T) {
		SetProcessModuleConfigNotifyAddress(ProcessConfigNotifyAddress{
			Process:       "monagent",
			NotifyAddress: "http://localhost:62222/api/v1/no/exist/route",
			Local:         false,
		})
		err := NotifyRemoteModuleConfig(context.Background(), &NotifyModuleConfig{
			Process: ProcessMonitorAgent,
			Module:  testFooModule,
			Config:  &Foo{Foo: "foofoo2", Bar: Bar{Bar: 2884}},
		})
		log.Errorf("err:%s", err)
		Convey("NotifyRemoteModuleConfig with err", t, func() {
			So(err, ShouldNotBeNil)
		})
	})
}

func TestNotifyConfigModule(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("notify config by modules (exist)", func(t *testing.T) {
		err := NotifyModules(context.Background(), []string{"foo"})
		Convey("NotifyModules", t, func() {
			So(err, ShouldBeNil)
		})
	})

	t.Run("notify config by modules (not exist)", func(t *testing.T) {
		noExistModuleReq := &NotifyModuleConfig{
			Module: "foo-not-exist",
		}
		err := notifyModuleConfig(context.Background(), noExistModuleReq)
		Convey("notifyModuleConfig", t, func() {
			So(err, ShouldNotBeNil)
		})

		err = NotifyLocalModuleConfig(context.Background(), noExistModuleReq)
		Convey("NotifyLocalModuleConfig", t, func() {
			So(err, ShouldNotBeNil)
		})
	})

}

func TestUpdateModuleConfigs(t *testing.T) {
	_init(t)
	defer cleanup()

	t.Run("update module configs by pair", func(t *testing.T) {
		err := UpdateConfigPairs(context.Background(), []string{"foo.foo=foo1"})
		Convey("UpdateConfigPairs", t, func() {
			So(err, ShouldBeNil)
		})
	})
}
