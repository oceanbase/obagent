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
