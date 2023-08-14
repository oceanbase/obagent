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
