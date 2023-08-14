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

package file

import (
	"context"

	"github.com/moby/sys/mountinfo"
)

func (f FileImpl) Mount(ctx context.Context, source, target, mType, options string) error {
	return mount(ctx, source, target, mType, options)
}

func (f FileImpl) Unmount(ctx context.Context, path string) error {
	return unmount(ctx, path)
}

func (f FileImpl) IsMountPoint(ctx context.Context, fileName string) (bool, error) {
	return mountinfo.Mounted(fileName)
}

func (f FileImpl) GetMountInfos(ctx context.Context, filter mountinfo.FilterFunc) ([]*mountinfo.Info, error) {
	return mountinfo.GetMounts(filter)
}
