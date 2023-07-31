//go:build linux
// +build linux

package agentd

import (
	"path/filepath"

	"github.com/containerd/cgroups"
	"github.com/opencontainers/runtime-spec/specs-go"
	log "github.com/sirupsen/logrus"
)

type LinuxLimiter struct {
	cgroup cgroups.Cgroup
}

func newLimiter(name string, conf LimitConfig) (Limiter, error) {
	if conf.CpuQuota <= 0 && conf.MemoryQuota <= 0 {
		log.Infof("create service %s resource limit skipped, no limit in config", name)
		return &LinuxLimiter{}, nil
	}
	cg, err := cgroups.New(cgroups.V1, cgroups.StaticPath(filepath.Join("/ocp_agent/", name)), toLinuxResources(conf))
	if err != nil {
		log.Warnf("create cgroup for service %s failed, fallback to watch limiter. only memory quota will affect!", name)
		return &WatchLimiter{
			name: name,
			conf: conf,
		}, nil
	}
	log.Infof("create service %s resource limit done, cpu: %v, memory: %v", name, conf.CpuQuota, conf.MemoryQuota)
	return &LinuxLimiter{
		cgroup: cg,
	}, nil
}

func (l *LinuxLimiter) LimitPid(pid int) error {
	if l.cgroup == nil {
		return nil
	}
	err := l.cgroup.Add(cgroups.Process{Pid: pid})
	return err
}

func toLinuxResources(conf LimitConfig) *specs.LinuxResources {
	var period *uint64 = nil
	var quota *int64 = nil
	var memLimit *int64 = nil

	if conf.CpuQuota > 0 {
		period = new(uint64)
		*period = 1000000
		quota = new(int64)
		*quota = int64(1000000 * conf.CpuQuota)
	}
	if conf.MemoryQuota > 0 {
		memLimit = new(int64)
		*memLimit = int64(conf.MemoryQuota)
	}
	return &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Period: period,
			Quota:  quota,
		},
		Memory: &specs.LinuxMemory{
			Limit: memLimit,
		},
	}
}
