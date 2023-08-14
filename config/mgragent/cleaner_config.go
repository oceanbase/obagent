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

package mgragent

import "time"

type Rule struct {
	// FileRegex Matches the regular expression of the file
	FileRegex string `json:"fileRegex" yaml:"fileRegex"`
	// RetentionDays Retention days
	RetentionDays uint64 `json:"retentionDays" yaml:"retentionDays"`
	// KeepPercentage Retention ratio, unit percentage, range: [0,100]
	KeepPercentage uint64 `json:"keepPercentage" yaml:"keepPercentage"`
}

type LogCleanerRules struct {
	// LogName Name of the log to be cleared
	LogName string `json:"logName" yaml:"logName"`
	// Path Log directory address
	Path string `json:"path" yaml:"path"`
	// DiskThreshold Disk clearing threshold (unit percentage) Range: [0,100]
	DiskThreshold uint64 `json:"diskThreshold" yaml:"diskThreshold"`
	// Rules Cleaning rule
	Rules []*Rule `json:"rules" yaml:"rules"`
}

// CleanerConfig Save the configuration of the cleaner
type CleanerConfig struct {
	LogCleaners []*LogCleanerRules `json:"logCleaners" yaml:"logCleaners"`
}

// ObCleaner Clear logs
type ObCleanerConfig struct {
	// RunInterval Running interval
	RunInterval time.Duration `json:"runInterval" yaml:"runInterval"`
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	// CleanerConf The configuration required to run
	CleanerConf *CleanerConfig `json:"cleanerConfig" yaml:"cleanerConfig"`
}
