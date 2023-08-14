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
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestVersion(t *testing.T) {
	log.SetLevel(logrus.DebugLevel)

	tempDir := _init(t)
	defer cleanup()

	configVersion1 := generateNewConfigVersion()
	snapshotPath1 := filepath.Join(tempDir, configVersion1.ConfigVersion)
	t.Run("save config snapshot version 1", func(t *testing.T) {
		os.RemoveAll(snapshotPath1)

		err := snapshotForConfigVersion(context.Background(), configVersion1.ConfigVersion)
		Convey("saveConfigSnapshot", t, func() {
			So(err, ShouldBeNil)
		})
	})

	time.Sleep(time.Millisecond)

	configVersion2 := generateNewConfigVersion()
	snapshotPath2 := filepath.Join(tempDir, configVersion2.ConfigVersion)
	t.Run("save config snapshot version 2", func(t *testing.T) {
		os.RemoveAll(snapshotPath2)

		err := snapshotForConfigVersion(context.Background(), configVersion2.ConfigVersion)
		Convey("saveConfigSnapshot", t, func() {
			So(err, ShouldBeNil)
		})
	})

	t.Run("rotate config version", func(t *testing.T) {
		err := checkConfigVersionBackups(context.Background(), 2, tempDir)
		Convey("rotate 2 with 2 versions", t, func() {
			So(err, ShouldBeNil)
		})

		err2 := checkConfigVersionBackups(context.Background(), 1, tempDir)
		Convey("rotate 1 with 2 versions", t, func() {
			So(err2, ShouldBeNil)

			_, version1Err := os.Stat(snapshotPath1)
			_, version2Err := os.Stat(snapshotPath2)
			So(version1Err, ShouldNotBeNil)
			So(version2Err, ShouldBeNil)
		})

	})

	t.Run("rotate config version", func(t *testing.T) {
		SetConfigMetaModuleConfigNotify(context.Background(), ConfigMetaBackup{
			MaxBackups: 0,
		})
		configMetaBackupWorker.checkOnce(context.Background())
		Convey("rotate 0 with 1 versions", t, func() {
			_, version1Err := os.Stat(snapshotPath1)
			_, version2Err := os.Stat(snapshotPath2)
			So(version1Err, ShouldNotBeNil)
			So(version2Err, ShouldNotBeNil)
		})
	})

}
