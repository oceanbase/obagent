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
