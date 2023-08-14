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

package api

import (
	"github.com/oceanbase/obagent/lib/http"
)

type Status struct {
	// agentd state
	State http.State `json:"state"`
	// whether agentd and all services are running
	Ready bool `json:"ready"`
	// agentd version
	Version string `json:"version"`
	// pid of agentd
	Pid int `json:"pid"`
	// socket file path
	Socket string `json:"socket"`
	// services (mgragent, monagent) status
	Services map[string]ServiceStatus `json:"services"`
	// services without agentd. maybe agentd dead, or service not exited expectedly
	Dangling []DanglingService `json:"dangling"`
	// StartAt is start time of agentd
	StartAt int64 `json:"startAt"`
}

type ServiceStatus struct {
	http.Status
	Socket string `json:"socket"`
	EndAt  int64  `json:"endAt"`
}

type DanglingService struct {
	Name    string `json:"name"`
	Pid     int    `json:"pid"`
	PidFile string `json:"pidFile"`
	Socket  string `json:"socket"`
}

type StartStopAgentParam struct {
	Service string `json:"service"`
}
