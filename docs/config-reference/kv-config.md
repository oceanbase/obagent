# kV 配置文件说明

本文介绍 KV 配置文件中的相关配置项，并列出了配置文件模板供您参考。

```yaml
# encrypted=true 的配置项，需要加密存储，目前仅支持 aes 加密。

## 基础认证相关
# basic_auth.yaml
configVersion: ""
configs:
    # agent http 接口用户名
    - key: agent.http.basic.auth.username
      value: ocp_agent
      valueType: string
    # agent http 接口密码
    - key: agent.http.basic.auth.password
      value: 
      valueType: string
      encrypted: true
    # 监控接口认证开关 
    - key: agent.http.basic.auth.metricAuthEnabled
      value: true
      valueType: string
      encrypted: false

## 元信息配置项
# common_meta.yaml
configVersion: ""
configs:
  # 配置最大备份版本数量
  - key: config.version.maxbackups
    value: 30
    valueType: int64
  # socks 代理开关
  - key: ocp.agent.http.socks.proxy.enabled
    value: false
    valueType: bool
  # socks 代理地址
  - key: ocp.agent.http.socks.proxy.address
    value: ""
    valueType: string
  # mgragent 进程端口
  - key: ocp.agent.manager.http.port
    value: 62888
    valueType: int64
  # monagent 进程端口
  - key: ocp.agent.monitor.http.port
    value: 62889
    valueType: int64

## agent日志相关
# log.yaml
configVersion: ""
configs:
  # monagent日志等级
  - key: monagent.log.level
    value: info
    valueType: string
  # monagent单个日志文件大小
  - key: monagent.log.maxsize.mb
    value: 200
    valueType: int64
  # monagent日志最大保留天数
  - key: monagent.log.maxage.days
    value: 30
    valueType: int64
  # monagent日志文件最大数量
  - key: monagent.log.maxbackups
    value: 15
    valueType: int64
  # monagent日志是否压缩
  - key: monagent.log.compress
    value: true
    valueType: bool
  # mgragent日志等级
  - key: mgragent.log.level
    value: info
    valueType: string
  # monagent单个日志文件大小
  - key: mgragent.log.maxsize.mb
    value: 200
    valueType: int64
  # monagent日志最大保留天数
  - key: mgragent.log.maxage.days
    value: 30
    valueType: int64
  # monagent日志文件最大数量
  - key: mgragent.log.maxbackups
    value: 15
    valueType: int64
  # mgragent日志是否压缩
  - key: mgragent.log.compress
    value: true
    valueType: bool

## observer日志清理相关
# ob_logcleaner.yaml
configVersion: ""
configs:
  # 是否清理 ob 日志
  - key: ob.logcleaner.enabled
    value: false
    valueType: bool
  # ob 日志清理间隔
  - key: ob.logcleaner.run.internal
    value: 5m
    valueType: string
  # ob 日志清理百分比阈值
  - key: ob.logcleaner.ob_log.disk.threshold
    value: 80
    valueType: int64
  # ob 日志清理0级规则之保留天数
  - key: ob.logcleaner.ob_log.rule0.retention.days
    value: 8
    valueType: int64
  # ob 日志清理0级规则之保留百分比
  - key: ob.logcleaner.ob_log.rule0.keep.percentage
    value: 60
    valueType: int64
  # ob 日志清理1级规则之保留天数
  - key: ob.logcleaner.ob_log.rule1.retention.days
    value: 30
    valueType: int64
  # ob 日志清理1级规则之保留百分比
  - key: ob.logcleaner.ob_log.rule1.keep.percentage
    value: 80
    valueType: int64
  # core 文件清理百分比阈值
  - key: ob.logcleaner.core_log.disk.threshold
    value: 80
    valueType: int64
  # core 文件清理保留天数
  - key: ob.logcleaner.core_log.rule0.retention.days
    value: 8
    valueType: int64
  # core 文件清理保留百分比
  - key: ob.logcleaner.core_log.rule0.keep.percentage
    value: 60
    valueType: int64

## 流水线相关
# monagent_pipeline.yaml
configVersion: ""
configs:
  # ob 监控用户
  - key: monagent.ob.monitor.user
    value: ocp_monitor
    valueType: string
  # ob 监控用户密码
  - key: monagent.ob.monitor.password
    value:
    valueType: string
    encrypted: true
  # ob sql 端口
  - key: monagent.ob.sql.port
    value: 2881
    valueType: int64
  # ob rpc 端口
  - key: monagent.ob.rpc.port
    value: 2882
    valueType: int64
  # 主机 ip
  - key: monagent.host.ip
    value: 127.0.0.1
    valueType: string
  # ob 集群名
  - key: monagent.ob.cluster.name
    value: ""
    valueType: string
  # ob 集群 id
  - key: monagent.ob.cluster.id
    value: 0
    valueType: int64
  # ob zone 名称
  - key: monagent.ob.zone.name
    value: ""
    valueType: string
  # ob 监控采集流水线开关
  - key: monagent.pipeline.ob.status
    value: inactive
    valueType: string
  # 主机监控采集流水线开关
  - key: monagent.pipeline.node.status
    value: active
    valueType: string
  # ob 日志路径
  - key: ob.log.path
    value: /data/log1
    valueType: string
  # ob 数据路径
  - key: ob.data.path
    value: /data/1
    valueType: string
  # ob 安装路径
  - key: ob.install.path
    value: /home/admin/oceanbase
    valueType: string
  # 主机只读挂载点
  - key: host.check.readonly.mountpoint
    value: /
    valueType: string
  # ob 日志采集流水线开关
  - key: monagent.pipeline.ob.log.status
    value: inactive
    valueType: string
  # es client 地址
  - key: es.client.addresses
    value: ""
    valueType: string
  # es client 用户名
  - key: es.client.auth.username
    value: ""
    valueType: string
  # es client 用户密码
  - key: es.client.auth.password
    value:
    valueType: string
    encrypted: true
  # ob 日志路径
  - key: observer.log.path
    value: /home/admin/oceanbase/log
    valueType: string
  # agent 日志路径
  - key: agent.log.path
    value: /home/admin/ocp_agent/log
    valueType: string
  # 系统日志路径
  - key: os.log.path
    value: /var/log
    valueType: string
  # 秒级监控采集间隔
  - key: monagent.second.metric.cache.update.interval
    value: 15s
    valueType: string
  # mysql 监控用户
  - key: monagent.mysql.monitor.user
    value: mysql_monitor_user
    valueType: string
  # mysql 监控用户名密码
  - key: monagent.mysql.monitor.password
    value:
    valueType: string
    encrypted: true
  # mysql sql 端口
  - key: monagent.mysql.sql.port
    value: 3306
    valueType: int64
  # mysql地址
  - key: monagent.mysql.host
    value: 127.0.0.1
    valueType: string
  # alertmanager 地址
  - key: monagent.alertmanager.address
    value:
    valueType: string
  # alertmanager 推送流水线开关
  - key: monagent.pipeline.ob.alertmanager.status
    value: inactive
    valueType: string
  # mysql 流水线开关
  - key: monagent.pipeline.mysql.status
    value: inactive
    valueType: string
```

## 配置模版

KV 的相关配置文件模板如下：

- basic_auth.yaml, 基础认证相关的 KV 配置项
- common_meta.yaml, 元信息的 KV 配置项
- log.yaml, agent日志相关的 KV 配置项
- ob_logcleaner.yaml, ob日志清理相关的 KV 配置项
- monagent_pipeline.yaml, 流水线相关的 KV 配置项
