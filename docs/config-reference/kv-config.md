# kV 配置文件说明

本文介绍 KV 配置文件中的相关配置项，并列出了配置文件模板供您参考。

```yaml
# encrypted=true 的配置项，需要加密存储，目前仅支持 aes 加密。
# 请将 {} 中的变量替换成您的真实值。如果您在 monagent 启动配置中将加密方法设置为 aes，您需要配置加密之后的值。

## 基础认证相关
# monagent_basic_auth.yaml
configVersion: "2021-08-20T07:52:28.5443+08:00"
configs:
    - key: http.server.basic.auth.username
      value: {http_basic_auth_user}
      valueType: string
    - key: http.server.basic.auth.password
      value: {http_basic_auth_password}
      valueType: string
      encrypted: true
    - key: http.admin.basic.auth.username
      value: {pprof_basic_auth_user}
      valueType: string
    - key: http.admin.basic.auth.password
      value: {pprof_basic_auth_password}
      valueType: string
      encrypted: true

## 流水线相关
# monagent_pipeline.yaml
configVersion: "2021-08-20T07:52:28.5443+08:00"
configs:
    - key: monagent.ob.monitor.user
      value: {monitor_user}
      valueType: string
    - key: monagent.ob.monitor.password
      value: {monitor_password}
      valueType: string
      encrypted: true
    - key: monagent.ob.sql.port
      value: {sql_port}
      valueType: int64
    - key: monagent.ob.rpc.port
      value: {rpc_port}
      valueType: int64
    - key: monagent.host.ip
      value: {host_ip}
      valueType: string
    - key: monagent.ob.cluster.name
      value: {cluster_name}
      valueType: string
    - key: monagent.ob.cluster.id
      value: {cluster_id}
      valueType: int64
    - key: monagent.ob.zone.name
      value: {zone_name}
      valueType: string
    - key: monagent.pipeline.ob.status
      value: {ob_monitor_status}
      valueType: string
    - key: monagent.pipeline.node.status
      value: {host_monitor_status}
      valueType: string

```

## 配置模版

KV 的相关配置文件模板如下：

- monagent_basic_auth.yaml，基础认证相关的 KV 配置项
- monagent_pipeline.yaml，流水线相关的 KV 配置项
