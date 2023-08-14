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
