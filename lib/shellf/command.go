package shellf

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/oceanbase/obagent/lib/shell"
)

type CommandParameterType string

const (
	packageNameType CommandParameterType = "PACKAGE_NAME"
)

// CommandGroup represents a group of commands that accomplish the same operation on different platforms.
// e.g. `rpm -i` and `dpkg -i` both belong to command group `install package`.
type CommandGroup struct {
	Name           string           // group name
	Branches       []*CommandBranch // branches, specified by the `case` node
	DefaultCommand *CommandTemplate // default command template, specified by the `default` node
}

// CommandBranch represents a single branch in CommandGroup that will be selected for a certain platform.
// The first branch that matches the os and arch of the current platform will be selected.
type CommandBranch struct {
	Os              string           // os, e.g. rhel, select condition
	Arch            string           // architecture, e.g. x86_64, select condition
	CommandTemplate *CommandTemplate // command template for this
}

// CommandTemplate is a template for a shell command.
// A command template has zero or more parameters.
// When provided with arguments, a command template can be instantiated to a real shell command.
type CommandTemplate struct {
	Template   string                      // command template with parameter placeholder, e.g. `rpm -qi ${PACKAGE_NAME}`
	Parameters map[string]CommandParameter // command parameters
	User       string                      // command execute user, e.g. admin
	Program    string                      // command execute program, e.g. sh
	Timeout    time.Duration               // command execute timeout, e.g. 30s
}

type CommandParameter struct {
	Name string               // command parameter name, corresponds to placeholder in command template
	Type CommandParameterType // command parameter type, for validation
}

// Instantiate Construct a command instance from this command template by replacing parameter placeholder. For example, with
// template `rpm -qi ${PACKAGE_NAME}` and args `{PACKAGE_NAME: oceanbase}`, we will get command `rpm -qi oceanbase`.
// However, if args does not contain the required `PACKAGE_NAME` argument, we will get an incomplete command with
// placeholder: `rpm -qi ${PACKAGE_NAME}`, no errors will be returned.
// Instantiate will check parameters first. If some arguments are invalid, no command will be constructed and an error
// will be returned.
func (t CommandTemplate) Instantiate(args map[string]string) (shell.Command, error) {
	for name, value := range args {
		if parameter, ok := t.Parameters[name]; ok {
			valid := parameter.Type.Validate(value)
			if !valid {
				return nil, errors.Errorf("invalid command argument, type: %s, value: %s", parameter.Type, value)
			}
		}
	}

	cmd := replaceAll(t.Template, args)
	command := shell.ShellImpl{}.NewCommand(cmd).WithOutputType(shell.StdOutput)
	if t.User != "" {
		command = command.WithUser(t.User)
	}
	if t.Program != "" {
		command = command.WithProgram(shell.Program(t.Program))
	}
	if t.Timeout > 0 {
		command = command.WithTimeout(t.Timeout)
	}
	return command, nil
}

func replaceAll(template string, args map[string]string) string {
	var oldnews []string
	for name, value := range args {
		key := "${" + name + "}"
		oldnews = append(oldnews, key, value)
	}
	replacer := strings.NewReplacer(oldnews...)
	return replacer.Replace(template)
}

func (b CommandBranch) String() string {
	var conditions []string
	if b.Os != "" {
		conditions = append(conditions, "os="+b.Os)
	}
	if b.Arch != "" {
		conditions = append(conditions, "arch="+b.Arch)
	}
	if len(conditions) == 0 {
		return "case(*)"
	} else {
		return "case(" + strings.Join(conditions, ", ") + ")"
	}
}

// Matches Check if this branch matches the os and arch of the current platform
// If os or arch of the current branch is empty, it will match any os or arch.
func (b CommandBranch) Matches(os string, arch string) bool {
	osMatches := b.Os == "" || b.Os == os
	archMatches := b.Arch == "" || b.Arch == arch
	return osMatches && archMatches
}

// SelectCommandTemplate Select a command template for a given os and arch.
// Try to match each branch in order.
// If no branches match, select the default command template.
func (g CommandGroup) SelectCommandTemplate(os string, arch string) *CommandTemplate {
	for _, branch := range g.Branches {
		if branch.Matches(os, arch) {
			return branch.CommandTemplate
		}
	}
	return g.DefaultCommand
}
