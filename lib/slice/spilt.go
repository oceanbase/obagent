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

package slice

import "github.com/pkg/errors"

func SpiltBatch(size, batchSize int, batchFn func(start, end int)) error {
	if batchSize <= 0 {
		return errors.Errorf("batchSize %d is invalid", batchSize)
	}
	times := size / batchSize
	for i := 0; i < times; i++ {
		batchFn(i*batchSize, (i+1)*batchSize)
	}
	if size%batchSize > 0 {
		batchFn(size/batchSize*batchSize, size)
	}

	return nil
}
