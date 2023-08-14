/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package cleaner

import (
	"context"
	"os"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config/mgragent"
	"github.com/oceanbase/obagent/errors"
)

var (
	obCleaner *ObCleaner
)

// ObCleaner clean observer log
type ObCleaner struct {
	*mgragent.ObCleanerConfig
	isStop chan bool
	// runCount Number of executions, there is a multiProcess race, requires atomic operations
	runCount uint64
}

func InitOBCleanerConf(conf *mgragent.ObCleanerConfig) error {
	return startOBCleaner(conf)
}

func UpdateOBCleanerConf(conf *mgragent.ObCleanerConfig) error {
	if obCleaner != nil && obCleaner.Enabled {
		// stop current cleaner
		obCleaner.Stop()
		obCleaner = nil
	}

	return startOBCleaner(conf)
}

// start ob cleaner by conf
func startOBCleaner(conf *mgragent.ObCleanerConfig) error {
	if obCleaner != nil {
		return errors.Errorf("ob cleaner has already been initialized.")
	}

	log.Infof("start ob cleaner")

	if !conf.Enabled {
		log.Warnf("ob cleaner is disabled")
		return nil
	}
	obCleaner = NewObCleaner(conf)
	go obCleaner.Run(context.Background())

	return nil
}

func NewObCleaner(conf *mgragent.ObCleanerConfig) *ObCleaner {
	if conf != nil && conf.RunInterval == 0 {
		conf.RunInterval = 300 * time.Second
	}
	return &ObCleaner{
		ObCleanerConfig: conf,
		isStop:          make(chan bool),
		runCount:        0,
	}
}

// DeleteFileByRetentionDays Delete the files whose modification date is earlier than the retention period
func (o *ObCleaner) DeleteFileByRetentionDays(ctx context.Context, dirToClean, fileRegex string, retentionDays uint64) error {
	ctxLog := log.WithContext(ctx)
	ctxLog.WithFields(log.Fields{
		"dirToClean":    dirToClean,
		"fileRegex":     fileRegex,
		"retentionDays": retentionDays,
	}).Info("delete file by retentionDays")

	realPath, err := GetRealPath(ctx, dirToClean)
	if err != nil {
		ctxLog.WithError(err).Error("GetRealPath failed")
		return err
	}

	ctxLog.WithField("realPath", realPath).Info("get real path")

	retentionDaysDuration := time.Duration(retentionDays) * time.Hour * 24
	matchedFiles, err := FindFilesByRegexAndMTime(ctx, realPath, fileRegex, retentionDaysDuration)
	if err != nil {
		ctxLog.WithError(err).Error("FindFilesByRegexAndMTime failed")
		return err
	}
	ctxLog.WithField("matchedFiles", matchedFiles).Info("get matched files")

	for _, fileInfo := range matchedFiles {
		err := os.Remove(fileInfo.Path)
		if err != nil {
			ctxLog.WithField("filePath", fileInfo.Path).WithError(err).Error("remove file failed")
			return err
		}
		ctxLog.WithField("filePath", fileInfo.Path).Info("delete file or dir")
	}

	return nil
}

// DeleteFileByKeepPercentage Delete files based on retention
func (o *ObCleaner) DeleteFileByKeepPercentage(ctx context.Context, dirToClean, fileRegex string, keepPercentage uint64) error {
	ctxLog := log.WithContext(ctx)
	ctxLog.WithFields(log.Fields{
		"dirToClean":     dirToClean,
		"fileRegex":      fileRegex,
		"keepPercentage": keepPercentage,
	}).Info("delete by keepPercentage")

	matchedFiles, err := FindFilesAndSortByMTime(ctx, dirToClean, fileRegex)
	if err != nil {
		ctxLog.WithError(err).Error("FindFilesAndSortByMTime failed")
		return err
	}
	ctxLog.WithField("matchedFiles", matchedFiles).Info("get matched(sorted) files")
	if len(matchedFiles) == 0 {
		return nil
	}
	var (
		totalSize       int64
		deletedSize     int64
		toBeDeletedSize int64
	)
	for _, file := range matchedFiles {
		totalSize += file.Info.Size()
	}
	toBeDeletedSize = int64(float64(totalSize) * ((100.0 - float64(keepPercentage)) / 100.0))
	for _, file := range matchedFiles {
		if deletedSize >= toBeDeletedSize {
			ctxLog.WithFields(log.Fields{
				"toBeDeletedSize": toBeDeletedSize,
				"deletedSize":     deletedSize,
				"totalSize":       totalSize,
			}).Info("delete files finished")
			break
		}
		err := os.Remove(file.Path)
		if err != nil {
			ctxLog.WithField("filePath", file.Path).WithError(err).Error("remove file failed")
			return err
		}

		ctxLog.WithFields(log.Fields{
			"filePath": file.Path,
			"fileSize": file.Info.Size(),
		}).Info("delete file")

		deletedSize += file.Info.Size()
	}

	return nil
}

func (o *ObCleaner) CleanFilesByRules(ctx context.Context, lcr *mgragent.LogCleanerRules) error {
	ctxLog := log.WithContext(ctx)
	ctxLog.WithField("logCleanerRules", lcr).Infof("run CleanFilesByRules task")

	usage, err := GetDiskUsage(ctx, lcr.Path)
	if err != nil {
		ctxLog.WithError(err).Error("GetDiskUsage failed")
		return err
	}
	ctxLog.Debugf("path %s usage %.2f before delete file by retention days", lcr.Path, usage)

	if usage <= float64(lcr.DiskThreshold) {
		ctxLog.Debugf("path %s usage %.2f %% is less than threshold %.2f %% ",
			lcr.Path, usage, float64(lcr.DiskThreshold))
		return nil
	}

	for _, rule := range lcr.Rules {
		err := o.DeleteFileByRetentionDays(ctx, lcr.Path, rule.FileRegex, rule.RetentionDays)
		if err != nil {
			ctxLog.WithFields(log.Fields{
				"path": lcr.Path,
				"rule": rule,
			}).WithError(err).Error("DeleteFileByRetentionDays failed")
			return err
		}
	}

	usage, err = GetDiskUsage(ctx, lcr.Path)
	if err != nil {
		ctxLog.WithError(err).Error("GetDiskUsage failed")
		return err
	}

	if usage <= float64(lcr.DiskThreshold) {
		ctxLog.Debugf("path %s usage %.2f %% is less than threshold %.2f %% after delete file by retention days",
			lcr.Path, usage, float64(lcr.DiskThreshold))
		return nil
	}

	ctxLog.Infof("path %s usage %.2f before delete file by keepPercentage", lcr.Path, usage)
	for _, rule := range lcr.Rules {
		err := o.DeleteFileByKeepPercentage(ctx, lcr.Path, rule.FileRegex, rule.KeepPercentage)
		if err != nil {
			ctxLog.WithFields(log.Fields{
				"path": lcr.Path,
				"rule": rule,
			}).WithError(err).Error("DeleteFileByKeepPercentage failed")
			return err
		}
	}
	return nil
}

func (o *ObCleaner) Clean(ctx context.Context) error {
	ctxLog := log.WithContext(ctx)
	if o.CleanerConf == nil {
		return nil
	}
	for _, logCleaner := range o.CleanerConf.LogCleaners {
		err := o.CleanFilesByRules(ctx, logCleaner)
		if err != nil {
			ctxLog.WithError(err).Errorf("clean %s failed", logCleaner.LogName)
			return err
		}
	}

	return nil
}

func (o *ObCleaner) Run(ctx context.Context) {
	ctxLog := log.WithContext(ctx)
	ticker := time.NewTicker(o.RunInterval)
	for {
		atomic.AddUint64(&o.runCount, 1)
		ctxLog.Infof("obCleaner run %d times", atomic.LoadUint64(&o.runCount))
		err := o.Clean(ctx)
		if err != nil {
			ctxLog.Error("run cleaning failed", err)
		}
		select {
		case <-ticker.C:
		case isStop := <-o.isStop:
			if isStop {
				ctxLog.Info("obCleaner finished")
				return
			}
		}
	}
}

func (o *ObCleaner) Stop() {
	log.Infof("stop ob cleaner")
	atomic.StoreUint64(&o.runCount, 0)
	o.isStop <- true
}
