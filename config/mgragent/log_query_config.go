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
