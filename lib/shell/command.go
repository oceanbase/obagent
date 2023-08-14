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

package shell

import (
	"context"
	"fmt"
	"time"

	"github.com/oceanbase/obagent/lib/mask"
)

type Program string
type OutputType string

const (
	Sh   Program = "sh"
	Bash Program = "bash"
	Zsh  Program = "zsh"
)

const (
	CombinedOutput OutputType = "combined"
	StdOutput      OutputType = "std"
)

const (
	RootUser  = "root"
	AdminUser = "admin"
)

const DefaultProgram = Sh
const DefaultOutputType = CombinedOutput
const DefaultTimeout = 10 * time.Second
const MaxTimeout = 30 * time.Minute // max half an hour
const MinTimeout = 1 * time.Second

type Command interface {
	Execute() (*ExecuteResult, error)
	ExecuteWithDebug() (*ExecuteResult, error)
	ExecuteAllowFailure() (*ExecuteResult, error)
	Cmd() string
	User() string
	Program() Program
	OutputType() OutputType
	Timeout() time.Duration
	WithUser(user string) Command
	WithProgram(program Program) Command
	WithOutputType(outputType OutputType) Command
	WithTimeout(timeout time.Duration) Command
	WithContext(ctx context.Context) Command
}

type command struct {
	user       string  // Run command as this user, if not provided, run command as current process's user
	program    Program // shell program to execute command, e.g. sh, bash
	outputType OutputType
	cmd        string
	timeout    time.Duration
	context    context.Context
}

func (c *command) Cmd() string {
	return c.cmd
}

func (c *command) User() string {
	return c.user
}

func (c *command) Program() Program {
	return c.program
}

func (c *command) OutputType() OutputType {
	return c.outputType
}

func (c *command) Timeout() time.Duration {
	return c.timeout
}

func (c *command) WithUser(user string) Command {
	c.user = user
	return c
}

func (c *command) WithProgram(program Program) Command {
	c.program = program
	return c
}

func (c *command) WithOutputType(outputType OutputType) Command {
	c.outputType = outputType
	return c
}

func (c *command) WithTimeout(timeout time.Duration) Command {
	c.timeout = adaptTimeout(timeout)
	return c
}

func (c *command) WithContext(ctx context.Context) Command {
	c.context = ctx
	return c
}

func (c *command) String() string {
	return fmt.Sprintf("Command{user=%s, program=%s, outputType=%s, cmd=%s, timeout=%s}", c.user, c.program, c.outputType, mask.Mask(c.cmd), c.timeout)
}

// adaptTimeout between MinTimeout and MaxTimeout
func adaptTimeout(timeout time.Duration) time.Duration {
	if timeout.Milliseconds() < MinTimeout.Milliseconds() {
		timeout = MinTimeout
	}
	if timeout.Milliseconds() > MaxTimeout.Milliseconds() {
		timeout = MaxTimeout
	}
	return timeout
}
