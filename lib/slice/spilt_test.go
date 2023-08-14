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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchValid(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	t.Run("batch_1", func(t *testing.T) {
		nums2 := make([]int, 0, len(nums))
		SpiltBatch(len(nums), 1, func(start, end int) {
			nums2 = append(nums2, nums[start:end]...)
		})
		assert.Equal(t, nums, nums2)
	})

	t.Run("batch_2", func(t *testing.T) {
		nums2 := make([]int, 0, len(nums))
		SpiltBatch(len(nums), 2, func(start, end int) {
			nums2 = append(nums2, nums[start:end]...)
		})
		assert.Equal(t, nums, nums2)
	})

	t.Run("batch_3", func(t *testing.T) {
		nums2 := make([]int, 0, len(nums))
		SpiltBatch(len(nums), 3, func(start, end int) {
			nums2 = append(nums2, nums[start:end]...)
		})
		assert.Equal(t, nums, nums2)
	})

	t.Run("batch size equal size", func(t *testing.T) {
		nums2 := make([]int, 0, len(nums))
		SpiltBatch(len(nums), len(nums), func(start, end int) {
			nums2 = append(nums2, nums[start:end]...)
		})
		assert.Equal(t, nums, nums2)
	})

	t.Run("batch bigger than size", func(t *testing.T) {
		nums2 := make([]int, 0, len(nums))
		SpiltBatch(len(nums), len(nums)+1, func(start, end int) {
			t.Log(start, end)
			nums2 = append(nums2, nums[start:end]...)
		})
		assert.Equal(t, nums, nums2)
	})
}

func TestBatchInvalid(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	t.Run("batch_0", func(t *testing.T) {
		err := SpiltBatch(len(nums), 0, func(start, end int) {
		})
		assert.NotNil(t, err)
	})
}
