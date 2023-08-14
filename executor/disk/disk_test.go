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

package disk

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/lib/system"
	"github.com/oceanbase/obagent/tests/mock"
)

func TestGetDiskUsage(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockDisk := mock.NewMockDisk(ctl)
	libDisk = mockDisk

	path := "/data/1"
	t.Run("get disk usage", func(t *testing.T) {
		mockDisk.EXPECT().GetDiskUsage(path).Return(&system.DiskUsage{}, nil)
		param := GetDiskUsageParam{Path: path}
		usage, err := GetDiskUsage(context.Background(), param)
		assert.Nil(t, err)
		assert.NotNil(t, usage)
	})
}
