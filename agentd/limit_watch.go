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

package agentd

import (
	"time"

	"github.com/shirou/gopsutil/v3/process"

	log "github.com/sirupsen/logrus"
)

type WatchLimiter struct {
	name string
	conf LimitConfig
}

func (l *WatchLimiter) LimitPid(pid int) error {
	if l.conf.MemoryQuota <= 0 {
		return nil
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			m, mErr := p.MemoryInfo()
			if mErr != nil {
				log.Warnf("fetch memory info for service %s failed: %v, stop watch loop", l.name, mErr)
				return // maybe process exited
			}
			if m.RSS > uint64(l.conf.MemoryQuota) {
				log.Warnf("service %s exceed the memory quota: %d, kill process %d", l.name, l.conf.MemoryQuota, p.Pid)
				_ = p.Kill()
			}
		}
	}()
	return nil
}
