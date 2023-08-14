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

package common

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/shell"
)

func GetMountPath(filePath string) string {
	_, err := os.Stat(filePath)
	if err != nil {
		log.Warnf("check filepath %s stat failed, err: %s", filePath, err)
		return ""
	}
	cmd := "df " + filePath
	command := shell.ShellImpl{}.NewCommand(cmd)
	result, err := command.ExecuteWithDebug()
	if err != nil {
		log.Warnf("get path's mount failed, filePath: %s", filePath)
		return ""
	}

	strs := strings.Split(result.Output, " ")
	mountedOn := strings.Replace(strs[len(strs)-1], "\n", "", -1)
	return mountedOn
}
