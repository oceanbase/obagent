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

package sdk

import (
	"context"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/lib/crypto"
	agentlog "github.com/oceanbase/obagent/log"
)

func init() {
	agentlog.InitLogger(agentlog.LoggerConfig{
		Level:      "debug",
		Filename:   "../tests/test.log",
		MaxSize:    10, // 10M
		MaxAge:     3,  // 3days
		MaxBackups: 3,
		LocalTime:  false,
		Compress:   false,
	})
}

var sdkOnce sync.Once

func initSDK() error {
	var err error
	sdkOnce.Do(func() {
		ctx := context.Background()
		err = InitSDK(ctx, config.SDKConfig{
			ConfigPropertiesDir: "../../etc/config_properties",
			ModuleConfigDir:     "../../etc/module_config",
			CryptoPath:          "../../etc/.config_secret.key",
			CryptoMethod:        crypto.PLAIN,
		})
		RegisterMgragentCallbacks(ctx)
		RegisterMonagentCallbacks(ctx)
	})

	return err
}

func TestInitSDK_Example(t *testing.T) {
	err := initSDK()
	assert.Nil(t, err)

	config.CurProcess = config.ProcessManagerAgent

	Convey("mgragent config", t, func() {

		common.InitBasicAuthConf(context.Background())

		err = config.NotifyModules(context.Background(), []string{config.ManagerAgentBasicAuthConfigModule})
		So(err, ShouldBeNil)
	})

	Convey("mgragent config", t, func() {
		err = config.InitModuleConfig(context.Background(), config.NotifyProcessConfigModule)
		So(err, ShouldBeNil)

		err = config.NotifyModules(context.Background(), []string{config.NotifyProcessConfigModule})
		So(err, ShouldBeNil)
	})
}

func TestInitSDK_WithWrongPath_Fail(t *testing.T) {
	Convey("init sdk with wrong config properties path", t, func() {
		err := InitSDK(context.Background(), config.SDKConfig{
			ConfigPropertiesDir: "../../etc/no-exist-path/config_properties",
			ModuleConfigDir:     "../../etc/module_config",
			CryptoPath:          "../../etc/.config_secret.key",
			CryptoMethod:        crypto.PLAIN,
		})
		So(err, ShouldNotBeNil)
	})

	Convey("init sdk with wrong module config path", t, func() {
		err := InitSDK(context.Background(), config.SDKConfig{
			ConfigPropertiesDir: "../../etc/config_properties",
			ModuleConfigDir:     "../../etc/no-exist-path/module_config",
			CryptoPath:          "../../etc/.config_secret.key",
			CryptoMethod:        crypto.PLAIN,
		})
		So(err, ShouldNotBeNil)
	})
}
func Test_processNotifyModuleConfigCallback(t *testing.T) {
	err := processNotifyModuleConfigCallback(context.Background(), nil)
	assert.NotNil(t, err)
}
