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

//go:build linux
// +build linux

package file

import (
	"context"

	mount_util "github.com/moby/sys/mount"
)

func mount(ctx context.Context, device, target, mType, options string) error {

	return mount_util.Mount(device, target, mType, options)
}

func unmount(ctx context.Context, path string) error {
	return mount_util.Unmount(path)
}
