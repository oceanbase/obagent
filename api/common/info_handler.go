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

package common

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/lib/system"
)

func TimeHandler(c *gin.Context) {
	SendResponse(c, time.Now(), nil)
}

func InfoHandler(c *gin.Context) {
	SendResponse(c, config.GetAgentInfo(), nil)
}

func GitInfoHandler(c *gin.Context) {
	SendResponse(c, config.GetGitInfo(), nil)
}

var StartAt = time.Now().UnixNano()
var libProcess system.Process = system.ProcessImpl{}

func StatusHandler(s *http.StateHolder) gin.HandlerFunc {
	return func(c *gin.Context) {
		ports := make([]int, 0)

		pid := os.Getpid()
		processInfo, err := libProcess.GetProcessInfoByPid(int32(pid))
		if err != nil {
			log.Errorf("StatusHandler get processInfo failed, pid:%s", pid)
		} else {
			ports = processInfo.Ports
		}
		var info = http.Status{
			State:   s.Get(),
			Version: config.AgentVersion,
			Pid:     pid,
			StartAt: StartAt,
			Ports:   ports,
		}
		SendResponse(c, info, nil)
	}
}
