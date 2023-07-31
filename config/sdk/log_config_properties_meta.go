package sdk

import (
	"github.com/oceanbase/obagent/config"
)

func setLogtailerConfigPropertyMeta() {
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logtailer.enabled",
			DefaultValue: "false",
			ValueType:    config.ValueBool,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logtailer.log.filter.rules.json.content",
			DefaultValue: "",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "logtailer.log.filter.rules.json.content",
			DefaultValue: "[]",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logtailer.keyword.alarm.rules",
			DefaultValue: "",
			ValueType:    config.ValueString,
		})

}

func setLogCleanerConfigPropertyMeta() {
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.enabled",
			DefaultValue: "false",
			ValueType:    config.ValueBool,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.run.internal",
			DefaultValue: "5m",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.ob_log.disk.threshold",
			DefaultValue: "80",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.ob_log.rule0.retention.days",
			DefaultValue: "8",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.ob_log.rule0.keep.percentage",
			DefaultValue: "60",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.ob_log.rule1.retention.days",
			DefaultValue: "30",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.ob_log.rule1.keep.percentage",
			DefaultValue: "80",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.core_log.disk.threshold",
			DefaultValue: "80",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.core_log.rule0.retention.days",
			DefaultValue: "8",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "ob.logcleaner.core_log.rule0.keep.percentage",
			DefaultValue: "60",
			ValueType:    config.ValueInt64,
		})
}

func setLogConfigPropertyMeta() {
	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.log.level",
			DefaultValue: "info",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.log.maxsize.mb",
			DefaultValue: "100",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.log.maxage.days",
			DefaultValue: "30",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.log.maxbackups",
			DefaultValue: "10",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "monagent.log.compress",
			DefaultValue: "true",
			ValueType:    config.ValueBool,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "mgragent.log.level",
			DefaultValue: "info",
			ValueType:    config.ValueString,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "mgragent.log.maxsize.mb",
			DefaultValue: "100",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "mgragent.log.maxage.days",
			DefaultValue: "30",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "mgragent.log.maxbackups",
			DefaultValue: "10",
			ValueType:    config.ValueInt64,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "mgragent.log.compress",
			DefaultValue: "true",
			ValueType:    config.ValueBool,
		})

	config.SetConfigPropertyMeta(
		&config.ConfigProperty{
			Key:          "config.version.maxbackups",
			DefaultValue: "30",
			ValueType:    config.ValueInt64,
		})
}
