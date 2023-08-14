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

package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func Fields(kv ...interface{}) *logrus.Entry {
	length := (len(kv) - 1)
	if length%2 == 1 {
		length -= 1
	}
	fields := logrus.Fields(make(map[string]interface{}, length))
	for i := 0; i < length<<1; i++ {
		fields[fmt.Sprint(kv[i>>1])] = kv[i>>1+1]
	}
	return logrus.WithFields(fields)
}
