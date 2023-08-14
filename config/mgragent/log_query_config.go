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

package mgragent

import "time"

type LogQueryConfig struct {
	// ErrCountLimit Upper limit for the total number of row errors in single-file processing logs
	ErrCountLimit       int                  `json:"errCountLimit" yaml:"errCountLimit"`
	QueryTimeout        time.Duration        `json:"queryTimeout" yaml:"queryTimeout"`
	DownloadTimeout     time.Duration        `json:"downloadTimeout" yaml:"downloadTimeout"`
	LogTypeQueryConfigs []LogTypeQueryConfig `json:"logTypeQueryConfigs" yaml:"logTypeQueryConfigs"`
}

type LogTypeQueryConfig struct {
	LogType string `json:"logType" yaml:"logType"`
	// IsOverrideByPriority Specifies whether logs of the ERROR level, WARN level,
	// and INFO level are used to override logs of the ERROR level.
	// For example, if INFO and ERROR are selected at the same time,
	// you only need to check the log file corresponding to INFO, and ERROR is also
	// included in the log file, so you do not need to check the two files
	IsOverrideByPriority    bool                     `json:"isOverrideByPriority" yaml:"isOverrideByPriority"`
	LogLevelAndFilePatterns []LogLevelAndFilePattern `json:"logLevelAndFilePatterns" yaml:"logLevelAndFilePatterns"`
}

type LogLevelAndFilePattern struct {
	LogLevel          string   `json:"logLevel" yaml:"logLevel"`
	Dir               string   `json:"dir" yaml:"dir"`
	FilePatterns      []string `json:"filePatterns" yaml:"filePatterns"`
	LogParserCategory string   `json:"logParserCategory" yaml:"logParserCategory"`
}
