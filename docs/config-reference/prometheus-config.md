# Prometheus 配置文件说明

本文介绍 Prometheus 配置文件中的相关配置项，并提供了配置文件模板供您参考。

## Prometheus 配置模板

`prometheus.yaml` 配置文件示例如下：

```yaml
# OBAgent 的 RPM 包中包含 Prometheus 的配置模版，您可以根据实际情况修改。
# 要开启基础认证，您需要配置 {http_basic_auth_user} 和 {http_basic_auth_password}。
# {target} 替换成监控目标主机的 IP 和 端口号, 如果有多个监控目标，需要配置多行，每个监控目标一行。
# rules 目录包含两个报警配置模版，分别是默认的主机和 OceanBase 数据库的报警配置。如需自定义报警项，您可以参考此目录。

# 全局配置
global:
  # 抓取间隔
  scrape_interval: 1s
  # 评估规则间隔
  evaluation_interval: 10s

# 报警规则配置
# Prometheus 将根据这些信息，推送报警信息至 alertmanager。
rule_files:
  - "rules/*rules.yaml"

# 抓取配置
# 用来配置 Prometheus 的数据采集。
scrape_configs:
  - job_name: prometheus
    metrics_path: /metrics
    scheme: http
    static_configs:
      - targets:
          - "localhost:9090"
  - job_name: node
    basic_auth:
      username: { http_basic_auth_user }
      password: { http_basic_auth_password }
    metrics_path: /metrics/node/host
    scheme: http
    static_configs:
      - targets:
          - { target }
  - job_name: ob_basic
    basic_auth:
      username: { http_basic_auth_user }
      password: { http_basic_auth_password }
    metrics_path: /metrics/ob/basic
    scheme: http
    static_configs:
      - targets:
          - { target }
  - job_name: ob_extra
    basic_auth:
      username: { http_basic_auth_user }
      password: { http_basic_auth_password }
    metrics_path: /metrics/ob/extra
    scheme: http
    static_configs:
      - targets:
          - { target }
  - job_name: agent
    basic_auth:
      username: { http_basic_auth_user }
      password: { http_basic_auth_password }
    metrics_path: /metrics/stat
    scheme: http
    static_configs:
      - targets:
          - { target }
```
