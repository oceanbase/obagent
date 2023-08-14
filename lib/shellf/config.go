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

// This file defines structures for decoding shellf config
package shellf

import (
	"io/ioutil"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ShellfConfig struct {
	CommandGroups []CommandGroupNode `yaml:"commandGroups"`
}

type CommandGroupNode struct {
	Name     string        `yaml:"name"`
	Program  string        `yaml:"program"`
	User     string        `yaml:"user"`
	Timeout  string        `yaml:"timeout"`
	Commands []CommandNode `yaml:"commands"`
	Params   []ParamNode   `yaml:"params"`
}

type CommandNode struct {
	Case    *CaseNode    `yaml:"case"`
	Default *DefaultNode `yaml:"default"`
	Cmd     string       `yaml:"cmd"`
	Program string       `yaml:"program"`
	User    string       `yaml:"user"`
	Timeout string       `yaml:"timeout"`
}

type CaseNode struct {
	Os   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type DefaultNode struct {
	// nothing
}

type ParamNode struct {
	Name     string `yaml:"name"`
	Validate string `yaml:"validate"`
}

func readCommandGroupMapFromFile(templatePath string) (map[string]*CommandGroup, error) {
	data, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, errors.Errorf("cannot read shell templates config file %s: %s", templatePath, err)
	}
	config := new(ShellfConfig)
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, errors.Errorf("cannot parse shell templates: %s", err)
	}

	groups, err := buildCommandGroupMap(config.CommandGroups)
	if err != nil {
		return nil, errors.Wrap(err, "build command group")
	}
	return groups, nil
}

func buildCommandGroupMap(nodes []CommandGroupNode) (map[string]*CommandGroup, error) {
	result := make(map[string]*CommandGroup)
	for _, node := range nodes {
		group, err := buildCommandGroup(node)
		if err != nil {
			return nil, errors.Wrap(err, "build command group map")
		}
		result[node.Name] = group
	}
	return result, nil
}

func buildCommandGroup(node CommandGroupNode) (*CommandGroup, error) {
	groupName := node.Name
	caseBranches, defaultCommand, err := buildBranches(node)
	if err != nil {
		return nil, errors.Wrap(err, "build command group")
	}
	return &CommandGroup{
		Name:           groupName,
		Branches:       caseBranches,
		DefaultCommand: defaultCommand,
	}, nil
}

func buildBranches(node CommandGroupNode) ([]*CommandBranch, *CommandTemplate, error) {
	var caseBranches []*CommandBranch
	var defaultCommand *CommandTemplate
	defaultCount := 0
	for _, commandNode := range node.Commands {
		if commandNode.Case != nil {
			caseNode := commandNode.Case
			branch := &CommandBranch{
				Os:              caseNode.Os,
				Arch:            caseNode.Arch,
				CommandTemplate: buildCommandTemplate(node, commandNode),
			}
			caseBranches = append(caseBranches, branch)
		} else {
			// We cannot recognize an empty `default` node, so if there is no `case` node,
			// consider it as default branch.
			defaultCommand = buildCommandTemplate(node, commandNode)
			defaultCount++
		}
	}
	if defaultCount != 1 {
		return nil, nil, errors.Errorf("invalid command group %s, one default command expected, %d actual", node.Name, defaultCount)
	}
	return caseBranches, defaultCommand, nil
}

func buildCommandTemplate(groupNode CommandGroupNode, commandNode CommandNode) *CommandTemplate {
	template := &CommandTemplate{
		Template:   commandNode.Cmd,
		Parameters: buildCommandParameters(groupNode.Params),
		Program:    groupNode.Program,
		User:       groupNode.User,
		Timeout:    parseTimeout(groupNode.Timeout),
	}
	// program, user and timeout specified in command node will override the one in group node
	if commandNode.Program != "" {
		template.Program = commandNode.Program
	}
	if commandNode.User != "" {
		template.User = commandNode.User
	}
	if commandNode.Timeout != "" {
		template.Timeout = parseTimeout(commandNode.Timeout)
	}
	return template
}

func parseTimeout(s string) time.Duration {
	if secs, err := strconv.Atoi(s); err == nil {
		return time.Duration(secs) * time.Second
	}
	if duration, err := time.ParseDuration(s); err == nil {
		return duration
	}
	return 0 * time.Second
}

func buildCommandParameters(nodes []ParamNode) map[string]CommandParameter {
	result := make(map[string]CommandParameter)
	for _, node := range nodes {
		parameter := CommandParameter{
			Name: node.Name,
			Type: CommandParameterType(node.Validate),
		}
		result[parameter.Name] = parameter
	}
	return result
}
