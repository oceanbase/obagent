package stat

import (
	"context"
)

type StatConfig struct {
	HostIP string `yaml:"host_ip"`
}

func SetStatConfig(ctx context.Context, conf StatConfig) error {
	HostIP = conf.HostIP
	RegisterStat(ctx)
	return nil
}
