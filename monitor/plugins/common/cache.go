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
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
)

type Cache struct {
	CacheStore sync.Map
	Cancel     context.CancelFunc
}

func (i *Cache) Close() {
	i.Cancel()
}

func (i *Cache) Update(ctx context.Context, key string, interval time.Duration, loadFunc func() (interface{}, error)) {
	metrics, err := loadFunc()
	if err != nil {
		log.WithContext(ctx).WithError(err).Warnf("update cache for key %s failed", key)
	}
	if metrics != nil {
		i.CacheStore.Store(key, metrics)
		log.Debugf("cache %s init success", key)
	}
	log.Infof("cache for %s updated firstly!", key)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			metrics, err = loadFunc()
			if err != nil {
				log.WithContext(ctx).WithError(err).Warnf("update cache for key %s failed", key)
			}
			if metrics != nil {
				i.CacheStore.Store(key, metrics)
				log.Debugf("cache %s update success", key)
			}
		case <-ctx.Done():
			return
		}
	}
}

type AllProcesses struct {
	Processes []*ProcessInfo
	Ctx       context.Context
	Cancel    context.CancelFunc
}

type ProcessInfo struct {
	Name     string
	Pid      int32
	UserName string
	Cmdline  string
}

var (
	allProcesses *AllProcesses
	processLock  sync.Mutex
)

func (p *AllProcesses) init() {
	process := make([]*ProcessInfo, 0)
	p.Processes = process
	p.refreshProcesses()
	p.Ctx, p.Cancel = context.WithCancel(context.Background())
	p.updateProcess()
}

func (p *AllProcesses) Close() {
	if p.Cancel != nil {
		p.Cancel()
	}
}

func GetProcesses() *AllProcesses {
	processLock.Lock()
	if allProcesses == nil {
		tmp := &AllProcesses{}
		tmp.init()
		allProcesses = tmp
	}
	processLock.Unlock()
	return allProcesses
}

func (p *AllProcesses) updateProcess() {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				p.refreshProcesses()
			case <-p.Ctx.Done():
				ticker.Stop()
				log.WithContext(p.Ctx).Infof("stop update process")
				return
			}
		}
	}()
}

func (p *AllProcesses) refreshProcesses() {
	processes, err := process.Processes()
	if err != nil {
		log.WithError(err).Warn("failed to list processes")
		return
	}
	processInfos := make([]*ProcessInfo, 0, len(processes))
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			log.WithError(err).Warnf("failed to get process name of pid: %d", proc.Pid)
			continue
		}
		username, err := proc.Username()
		if err != nil {
			log.WithError(err).Warnf("failed to get process username of pid: %d", proc.Pid)
			continue
		}
		cmdline, err := proc.Cmdline()
		if err != nil {
			log.WithError(err).Warnf("failed to get process cmdline of pid: %d", proc.Pid)
			continue
		}
		processInfos = append(processInfos, &ProcessInfo{Name: name, Pid: proc.Pid, UserName: username, Cmdline: cmdline})
	}
	p.Processes = processInfos
}
