# 配置模板文件说明

配置模板中定义了各个模块的配置模板，以及各个监控的流水线。这里面的大部分配置项都是用了${config_key}来进行表示，${config_key}和config_properties下KV-config关联，并在启动agent时进行替换。
配置模板文件的说明见下表：

配置文件名称 | 说明
--- | ---
common_module.yaml | monagent和mgragent meta信息配置模块。
log_module.yaml | mongagent和mgragent 日志相关配置。
mgragent_logquerier_module.yaml | ob和agent日志采集相关配置。
mgragent_module.yaml | mgragent认证和配置更新相关配置。
monagent_basic_auth.yaml | monagent认证相关配置。
monitor_host_log.yaml | 主机日志采集推送ES流水线配置模板。
monitor_mysql.yaml | mysql监控采集流水线配置模板。
monitor_node_host.yaml | 主机监控采集流水线配置模板。 
monitor_ob.yaml | ob性能监控采集流水线配置模板。
monitor_ob_custom.yaml | ob连接和进程监控流水线配置模板。
monitor_ob_log.yaml | ob error日志采集流水线配置模板。
monitor_observer_log.yaml | ob日志采集推送ES流水线配置模板。
ob_logcleaner_module.yaml | ob日志清理模块配置模板。

