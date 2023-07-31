package process

import (
	"time"
)

// ProcessConfig a process config. defines how to run a process
type ProcessConfig struct {
	Program string
	Args    []string

	Cwd        string
	User       string
	Group      string
	Stdout     string
	Stderr     string
	InheritEnv bool
	Envs       map[string]string
	Rlimit     map[string]int64

	KillWait  time.Duration
	FinalWait time.Duration
}
