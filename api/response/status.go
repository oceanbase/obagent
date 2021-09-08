// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package response

import "encoding/json"

type State int32

const (
	Unknown     State = 0
	Starting    State = 1
	Running     State = 2
	Stopping    State = 3
	Stopped     State = 4
	StoppedWait State = 5
)

var stateString = []string{"unknown", "starting", "running", "stopping", "stopped", "stopped_wait"}

func (s State) String() string {
	return stateString[int(s)]
}

func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *State) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	for i := 0; i < len(stateString); i++ {
		if stateString[i] == str {
			*s = State(i)
			return nil
		}
	}
	*s = Unknown
	return nil
}

//Status info api response
type Status struct {
	//service state
	State State `json:"state"`
	//service version
	Version string `json:"version"`
	//service pid
	Pid int `json:"pid"`
	//timestamp when service started
	StartAt int64 `json:"startAt"`
}
