//go:build darwin
// +build darwin

package file

import "context"

func mount(ctx context.Context, device, target, mType, options string) error {
	panic("mount not supported on this platform")
}

func unmount(ctx context.Context, path string) error {
	panic("mount not supported on this platform")
}
