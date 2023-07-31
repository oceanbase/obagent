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
