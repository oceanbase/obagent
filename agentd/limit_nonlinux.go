//go:build !linux
// +build !linux

package agentd

import (
	log "github.com/sirupsen/logrus"
)

func newLimiter(name string, conf LimitConfig) (Limiter, error) {
	log.Infof("creating service %s resource limit done, WatchLimiter memory: %v", name, conf.MemoryQuota)
	return &WatchLimiter{name: name, conf: conf}, nil
}
