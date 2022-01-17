OBAgent 集成了 Prometheus 提供的 mysqld exporter, 可以使用 OBAgent 来采集 MySQL 的性能指标。
​

# 配置方式
OBAgent 中打包了采集 MySQL 的流水线配置，默认没有开启，可以通过修改如下配置项来配置并开启 MySQL 的采集。
```yaml
# 用户采集监控数据的 MySQL 用户, 对于开启的采集项，需要有对应表的读权限
- key: monagent.mysql.monitor.user
  value: mysql_monitor_user
  valueType: string
# 监控用户的密码
- key: monagent.mysql.monitor.password
  value: mysql_monitor_password
  valueType: string
  encrypted: true
# MySQL 的 SQL 端口
- key: monagent.mysql.sql.port
  value: 3306
  valueType: int64
# MySQL 的连接地址
- key: monagent.mysql.host
  value: 127.0.0.1
  valueType: string
# 是否开启 MySQL 指标采集，默认为 inactive 表示不开启，如果开启需要修改为 active
- key: monagent.pipeline.mysql.status
  value: inactive
  valueType: string
```


MySQL 流水线的配置如下，采集的指标保持 mysqld exporter 的默认开关，一般不需要修改，如果需要指定特定指标的开关，需要修改 mysqldInput 插件 scraperFlags 的配置。
```yaml
# 当前配置文件在 OBAgent 的 RPM 包中， monitor_mysql.yaml, 这里只展示采集插件部分

mysqldInput: &mysqldInput
  plugin: mysqldInput
  config:
    timeout: 10s
    pluginConfig:
      dsn: ${monagent.mysql.monitor.user}:${monagent.mysql.monitor.password}@(${monagent.mysql.host}:${monagent.mysql.sql.port})/
      scraperFlags:
        # 开启 binlog_size 指标的采集
        binlog_size: true
        # 关闭 slave_status 指标的采集
        slave_status: false

```

> **注意：** MySQL 相关的采集配置未与 OBD 集成，只能以手动配置的方式来启动。
