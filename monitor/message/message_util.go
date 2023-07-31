package message

import (
	log "github.com/sirupsen/logrus"
)

func UniqueMetrics(metrics []*Message) []*Message {
	existsMap := make(map[string]bool)
	ret := make([]*Message, 0, len(metrics))
	for _, entry := range metrics {
		id := entry.Identifier()
		_, exists := existsMap[id]
		if exists {
			log.Debugf("message conflict with existing, message: %v", entry)
			continue
		}
		existsMap[id] = true
		ret = append(ret, entry)
	}
	return ret
}
