package host

import (
	"context"
	"math"
	"net"
	"os/exec"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shirou/gopsutil/v3/cpu"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/shell"
	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/monitor/message"
	"github.com/oceanbase/obagent/monitor/plugins/common"
	"github.com/oceanbase/obagent/monitor/utils"
)

const sampleConfig = `
timeout: 10s
`

const description = `
custom logic for internal use
`

type SmartLogInfo struct {
	SmartAttributeName string `json:"smart_attribute_name"`
	SmartValue         string `json:"smart_value"`
}

type SmartLogResponse struct {
	SmartLogInfoList []SmartLogInfo `json:"smartlog_info"`
}

type CustomInputConfig struct {
	Timeout  time.Duration `yaml:"timeout"`
	Interval time.Duration `yaml:"collect_interval"`
}

type CustomInput struct {
	Config   *CustomInputConfig
	LibShell shell.Shell
	Env      string

	ctx  context.Context
	done chan struct{}

	tcpOutSegs, tcpRetrans int64
	devIoTicks             map[string]float64
	devIoAwaits            map[string]map[string]float64
}

func (c *CustomInput) Init(ctx context.Context, config map[string]interface{}) error {
	var pluginConfig CustomInputConfig
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "process input encode config")
	}
	err = yaml.Unmarshal(configBytes, &pluginConfig)
	if err != nil {
		return errors.Wrap(err, "process input decode config")
	}
	c.Config = &pluginConfig
	c.LibShell = shell.ShellImpl{}
	c.ctx = ctx
	c.done = make(chan struct{})

	env, err := common.CheckNodeEnv(ctx)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("check node env failed")
	} else {
		c.Env = env
	}

	// init tcpRetrans
	c.tcpOutSegs, c.tcpRetrans, err = gatherTcpretransRadio(ctx)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("gather TcpRetransSegs and TcpOutSegs failed")
	}
	// init devIoTicks and gatherSeconds
	c.devIoTicks, c.devIoAwaits, err = gatherIoTicks(ctx)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("gather devIoTicks failed")
	}

	return nil
}

func (c *CustomInput) SampleConfig() string {
	return sampleConfig
}

func (c *CustomInput) Description() string {
	return description
}

func (c *CustomInput) Start(out chan<- []*message.Message) error {
	log.WithContext(c.ctx).Infof("start customInput plugin")
	go c.update(out)
	return nil
}

func (c *CustomInput) update(out chan<- []*message.Message) {
	ticker := time.NewTicker(c.Config.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			msgs, err := c.CollectMsgs(c.ctx)
			if err != nil {
				log.WithError(err).Warn("customInfo collect failed")
				continue
			}
			out <- msgs
		case <-c.done:
			log.Info("customInput plugin exited")
			return
		}
	}
}

func (c *CustomInput) Stop() {
	if c.done != nil {
		close(c.done)
	}
}

func (c *CustomInput) doCollectNtpOffset(ctx context.Context) *message.Message {
	var metricEntry *message.Message
	ntpdIsExist := checkNtpProcess("ntpd", ctx)
	if ntpdIsExist {
		cmd := "ntpq -p"
		command := c.LibShell.NewCommand(cmd).WithContext(ctx).WithOutputType(shell.StdOutput).WithTimeout(c.Config.Timeout)
		result, err := command.ExecuteWithDebug()
		if err == nil {
			offset, err := processNtpqOutput(result.Output)
			if err != nil {
				log.WithContext(ctx).Warnf("process ntpq output failed, err: %s", err)
				return metricEntry
			}
			metricEntry = message.NewMessage("node_ntp_offset_seconds", message.Gauge, time.Now()).
				AddTag("env_type", c.Env).
				AddField("value", offset)
		}
		return metricEntry
	}

	chronydIsExist := checkNtpProcess("chronyd", ctx)
	if chronydIsExist {
		chronycPath, err := exec.LookPath("chronyc")
		if chronycPath != "" && err == nil {
			cmd := chronycPath + " tracking -n"
			command := c.LibShell.NewCommand(cmd).WithContext(ctx).WithOutputType(shell.StdOutput).WithTimeout(c.Config.Timeout)
			result, err := command.ExecuteWithDebug()
			if err == nil {
				offset, err := processChronycOutput(result.Output)
				if err != nil {
					log.WithContext(ctx).Warnf("process chronyc output failed, reason: %s", err)
					return metricEntry
				}
				metricEntry = message.NewMessage("node_ntp_offset_seconds", message.Gauge, time.Now()).
					AddTag("env_type", c.Env).
					AddField("value", offset)
			} else {
				log.WithContext(ctx).Warnf("get chronyc info failed, reason: %s", err)
			}
		} else {
			log.WithContext(ctx).Warnf("couldn't found chronyc, reason: %s", err)
		}
		return metricEntry
	}

	if c.Env == common.Container {
		log.WithContext(ctx).Debug("the current environment is container, skip collect node_ntp_offset_seconds")
		return metricEntry
	}
	log.WithContext(ctx).Warn("couldn't found ntpd or chronyd in host")
	return metricEntry
}

func (c *CustomInput) doCollectTcpRetrans(ctx context.Context) *message.Message {
	var err error
	var tcpRetransRatio float64
	preRetrans, preOutSegs := c.tcpRetrans, c.tcpOutSegs
	c.tcpOutSegs, c.tcpRetrans, err = gatherTcpretransRadio(ctx)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("failed to gather current TcpRetransSegs and TcpOutSegs")
	}
	outSegsDiff := c.tcpOutSegs - preOutSegs
	if outSegsDiff == 0 {
		tcpRetransRatio = 0.0
	} else {
		tcpRetransRatio = float64((c.tcpRetrans-preRetrans)*100) / float64(outSegsDiff)
	}
	var metricEntry *message.Message
	tcpRetransRatio = math.Trunc(tcpRetransRatio*100) / 100
	metricEntry = message.NewMessage("tcp_retrans", message.Gauge, time.Now()).
		AddTag("env_type", c.Env).
		AddField("value", tcpRetransRatio)

	return metricEntry
}

func (c *CustomInput) doCollectIoInfos(ctx context.Context) []*message.Message {
	var err error
	ioUtils := make(map[string]float64)
	ioAwaits := make(map[string]float64)
	preIoutils := make(map[string]float64)
	for k, v := range c.devIoTicks {
		preIoutils[k] = v
	}
	preDevIoAwaits := make(map[string]map[string]float64)
	for k, v := range c.devIoAwaits {
		preDevIoAwaits[k] = v
	}
	c.devIoTicks, c.devIoAwaits, err = gatherIoTicks(ctx)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("failed to gather ioUtils info")
	}
	for dev, value := range c.devIoTicks {
		if v, ok := preIoutils[dev]; ok {
			ioUtils[dev] = (value - v) / float64(60*10)
		}
	}
	for dev, value := range c.devIoAwaits {
		if preValue, ok := preDevIoAwaits[dev]; ok {
			rdIos := value["readCount"] - preValue["readCount"]
			wrIos := value["writeCount"] - preValue["writeCount"]
			rdTicks := value["readTicks"] - preValue["readTicks"]
			wrTicks := value["writeTicks"] - preValue["writeTicks"]
			nIos := rdIos + wrIos
			nTicks := rdTicks + wrTicks

			var await float64
			if nIos == 0.0 {
				await = 0.0
			} else {
				await = math.Trunc(100*nTicks/nIos) / 100
			}
			ioAwaits[dev] = await
		}
	}
	metrics := make([]*message.Message, 0)
	var metricEntry *message.Message
	for dev, value := range ioUtils {
		value = math.Trunc(value*100) / 100
		metricEntry = message.NewMessage("io_util", message.Gauge, time.Now()).
			AddTag("env_type", c.Env).
			AddTag("dev", dev).
			AddField("value", value)
		metrics = append(metrics, metricEntry)
	}
	for dev, value := range ioAwaits {
		var messageEntry *message.Message
		messageEntry = message.NewMessage("io_await", message.Gauge, time.Now()).
			AddTag("env_type", c.Env).
			AddTag("dev", dev).
			AddField("value", value)
		metrics = append(metrics, messageEntry)
	}

	return metrics
}

func (c *CustomInput) doCollectCPUCount(ctx context.Context) *message.Message {
	var metricEntry *message.Message
	entry := log.WithContext(context.WithValue(ctx, agentlog.StartTimeKey, time.Now()))
	count, err := cpu.Counts(true)
	entry.Debug("collect cpu_count end")
	if err != nil {
		log.WithContext(ctx).Warn("failed to get cpu count")
	} else {
		v, ok := utils.ConvertToFloat64(count)
		if !ok {
			log.WithContext(ctx).WithError(err).Warn("failed to convert cpu count")
		} else {
			metricEntry = message.NewMessage("cpu_count", message.Gauge, time.Now()).
				AddTag("env_type", c.Env).
				AddField("value", v)
		}
	}
	return metricEntry
}

func (c *CustomInput) doCollectBandwidthInfo(ctx context.Context) []*message.Message {
	var metricEntrys = make([]*message.Message, 0)
	infos, err := net.Interfaces()
	if err != nil {
		log.WithContext(ctx).Warn("failed to get network interfaces info")
	}
	for _, info := range infos {
		if info.Flags > 0 && info.Flags%2 == 1 {
			cmd := "ethtool " + info.Name
			command := c.LibShell.NewCommand(cmd).WithContext(ctx).WithTimeout(c.Config.Timeout)
			result, err := command.ExecuteWithDebug()
			if err != nil {
				log.WithContext(ctx).Warnf("get bandwidth of net interface %s failed, cmd: %s", info.Name, cmd)
				continue
			}
			var metricEntry *message.Message
			if err == nil && result.ExitCode == 0 {
				speed := parseEthtoolResult(result.Output)
				metricEntry = message.NewMessage("node_net_bandwidth_bps", message.Gauge, time.Now()).
					AddTag("env_type", c.Env).
					AddTag("device", info.Name).
					AddField("value", speed)
				metricEntrys = append(metricEntrys, metricEntry)
			}
		}
	}
	return metricEntrys
}

func (c *CustomInput) CollectMsgs(ctx context.Context) ([]*message.Message, error) {
	metrics := make([]*message.Message, 0, 4)
	ioInfos := c.doCollectIoInfos(ctx)
	cpuCountInfo := c.doCollectCPUCount(ctx)
	bandwidthInfo := c.doCollectBandwidthInfo(ctx)
	ntpInfo := c.doCollectNtpOffset(ctx)
	if ntpInfo != nil {
		metrics = append(metrics, ntpInfo)
	}
	if ioInfos != nil {
		metrics = append(metrics, ioInfos...)
	}
	if cpuCountInfo != nil {
		metrics = append(metrics, cpuCountInfo)
	}
	if bandwidthInfo != nil {
		metrics = append(metrics, bandwidthInfo...)
	}

	return metrics, nil
}
