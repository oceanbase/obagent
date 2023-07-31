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
