package sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/config"
)

func TestLogModule(t *testing.T) {
	err := initSDK()
	assert.Nil(t, err)

	ctx := context.Background()
	// init log
	err = config.InitModuleConfig(ctx, config.ManagerLogConfigModule)
	assert.Nil(t, err)

	// init basic auth
	common.InitBasicAuthConf(ctx)

	// init notify process
	err = config.InitModuleConfig(ctx, config.NotifyProcessConfigModule)
	assert.Nil(t, err)

	err = config.InitModuleConfig(ctx, config.OBLogcleanerModule)
	assert.Nil(t, err)

}
