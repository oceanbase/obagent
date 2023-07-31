package system

import (
	"context"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/system"
)

var libSystem system.System = system.SystemImpl{}

func GetHostInfo(ctx context.Context) (*system.HostInfo, *errors.OcpAgentError) {
	info, err := libSystem.GetHostInfo()
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return info, nil
}
