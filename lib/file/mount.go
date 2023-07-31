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
