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

package http

import (
	"encoding/json"
	"sync/atomic"
)

type State int32

const (
	Unknown  State = 0
	Starting State = 1
	Running  State = 2
	Stopping State = 3
	Stopped  State = 4
)

var stateString = []string{"unknown", "starting", "running", "stopping", "stopped"}

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

// Status info api response
type Status struct {
	//service state
	State State `json:"state"`
	//service version
	Version string `json:"version"`
	//service pid
	Pid int `json:"pid"`
	//timestamp when service started
	StartAt int64 `json:"startAt"`
	// Ports process occupied ports
	Ports []int `json:"ports"`
}

type StateHolder struct {
	state State
}

func NewStateHolder(init State) *StateHolder {
	return &StateHolder{
		state: init,
	}
}

func (s *StateHolder) Get() State {
	return State(atomic.LoadInt32((*int32)(&s.state)))
}

func (s *StateHolder) Set(new State) {
	atomic.StoreInt32((*int32)(&s.state), int32(new))
}

func (s *StateHolder) Cas(old, new State) bool {
	return atomic.CompareAndSwapInt32((*int32)(&s.state), int32(old), int32(new))
}
