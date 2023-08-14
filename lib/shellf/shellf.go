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

package shellf

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/lib/shell"
	"github.com/oceanbase/obagent/lib/system"
)

type ShellShelf interface {
	GetCommandForCurrentPlatform(commandGroupName string, args map[string]string) (shell.Command, error)
	GetCommand(commandGroupName string, os string, arch string, args map[string]string) (shell.Command, error)
}

type shellShelf struct {
	groups *map[string]*CommandGroup // groupName -> CommandGroup map
}

func (s shellShelf) GetCommandForCurrentPlatform(commandGroupName string, args map[string]string) (shell.Command, error) {
	hostInfo, err := system.SystemImpl{}.GetHostInfo()
	if err != nil {
		log.WithError(err).Info("get command from shelf, failed to get host info, use default command")
		return s.GetCommand(commandGroupName, "", "", args)
	}
	return s.GetCommand(commandGroupName, hostInfo.OsPlatformFamily, hostInfo.Architecture, args)
}

func (s shellShelf) GetCommand(commandGroupName string, os string, arch string, args map[string]string) (shell.Command, error) {
	var configFile = path.ConfDir() + "/shell_templates/shell_template.yaml"

	err := s.init(configFile)
	if err != nil {
		log.WithError(err).Warn("init shellf failed")
	}
	commandGroup, ok := (*s.groups)[commandGroupName]
	if !ok {
		return nil, errors.Errorf("command group %s not found", commandGroupName)
	}
	commandTemplate := commandGroup.SelectCommandTemplate(os, arch)
	command, err := commandTemplate.Instantiate(args)
	if err != nil {
		return nil, errors.Wrap(err, "get command from shelf")
	}
	log.Infof("get command from shelf, os=%v, arch=%v, cmd=%v", os, arch, command.Cmd())
	return command, nil
}

func (s *shellShelf) init(templatePath string) error {
	if len(*s.groups) > 0 {
		return nil
	}
	groups, err := readCommandGroupMapFromFile(templatePath)
	if err != nil {
		return errors.Wrap(err, "init shellf")
	}
	for k, v := range groups {
		(*s.groups)[k] = v
	}
	return nil
}

var shelf = (func() shellShelf {
	groups := make(map[string]*CommandGroup)
	return shellShelf{groups: &groups}
})()

func Shelf() ShellShelf {
	return shelf
}

func InitShelf(templatePath string) {
	err := shelf.init(templatePath)
	if err != nil {
		log.WithError(err).Fatal("init shellf failed")
	}
}
