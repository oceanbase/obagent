package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/lib/shell"
	"github.com/oceanbase/obagent/monitor/utils"
)

func TestInitCustomInput(t *testing.T) {
	config := `
        timeout: 10s
        interval: 10s
        ntp_server: 127.0.0.1
    `
	var configMap map[string]interface{}
	yaml.Unmarshal([]byte(config), &configMap)
	customInput := &CustomInput{}
	err := customInput.Init(context.Background(), configMap)
	defer customInput.Stop()
	require.True(t, err == nil)
}

func TestCollectTcpRetrans(t *testing.T) {
	config := `
        timeout: 10s
        interval: 10s
    `
	customInput := getCustomInput(config)

	metricEntry := customInput.doCollectTcpRetrans(context.Background())
	_, ok := metricEntry.GetField("value")
	require.True(t, ok)
}

func TestCollectIoUtils(t *testing.T) {
	config := `
        timeout: 10s
    `
	customInput := getCustomInput(config)

	metrics := customInput.doCollectIoInfos(context.Background())
	require.True(t, len(metrics) >= 0)
}

func TestCollectCpuCountInfo(t *testing.T) {
	config := `
        timeout: 10s
    `
	customInput := getCustomInput(config)

	metricEntry := customInput.doCollectCPUCount(context.Background())
	v, found := metricEntry.GetField("value")
	require.True(t, found)
	value, ok := utils.ConvertToFloat64(v)
	require.True(t, ok)
	require.True(t, value > 0)
}

func TestCollectBandwidthInfo(t *testing.T) {
	config := `
        timeout: 10s
    `
	customInput := getCustomInput(config)

	metrics := customInput.doCollectBandwidthInfo(context.Background())
	for _, metric := range metrics {
		_, foundFiled := metric.GetField("value")
		require.True(t, foundFiled)
		_, foundTag := metric.GetTag("device")
		require.True(t, foundTag)
	}
}

func getCustomInput(config string) *CustomInput {
	var configMap map[string]interface{}
	yaml.Unmarshal([]byte(config), &configMap)
	var pluginConfig CustomInputConfig
	configBytes, _ := yaml.Marshal(config)
	yaml.Unmarshal(configBytes, &pluginConfig)

	customInput := &CustomInput{}
	customInput.Config = &pluginConfig
	customInput.LibShell = shell.ShellImpl{}

	return customInput
}
