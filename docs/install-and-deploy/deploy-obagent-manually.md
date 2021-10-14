# 手动部署 OBAgent

OBAgent 提供使用 OBD 部署和手动部署。要手动部署 OBAgent，您要配置 OBAgent、Prometheus 和 Prometheus Alertmanager（可选）。推荐您使用 OBD 部署 OBAgent。

## 前提条件

在部署 OBAgent 之前，您需要确认以下信息：

- OceanBase 数据库服务已经部署并启动。
- OBAgent 已经安装。详细信息，参考 [安装 OBAgent](install-obagent.md)。
- OBAgent 的默认端口 8088、8089 未占用。您也可以自定义端口。

## 操作步骤

按以下步骤部署 OBAgent：

### 步骤1：部署 monagent

1. 修改配置文件，详细信息，参考 [monagent 配置](../config-reference/monagent-config.md) 和 [KV 配置](../config-reference/kv-config.md)。

2. 启动 monagent 进程。推荐您使用 Supervisor 启动 monagent 进程。

    ```bash
    # 将当前目录切换至 OBAgent 的安装目录
    cd /home/admin/obagent
    # 启动 monagent 进程
    nohup ./bin/monagent -c conf/monagent.yaml >> ./log/monagent_stdout.log 2>&1 &

    ```

    ```bash
    # Supervisor 配置样例
    [program:monagent]
    command=./bin/monagent -c conf/monagent.yaml
    directory=/home/admin/obagent
    autostart=true
    autorestart=true
    redirect_stderr=true
    priority=10
    stdout_logfile=log/monagent_stdout.log
    ```

### （可选）步骤2：部署 Prometheus

> 说明：您需要安装 Prometheus。

1. 配置 Prometheus，详情参考 [Prometheus 配置文件说明](../config-reference/prometheus-config.md)。
2. 启动 Prometheus。

    ```bash
    ./prometheus --config.file=./prometheus.yaml
    ```

### （可选）步骤3：部署 Prometheus Alertmanager

- 下载并解压 Prometheus Alertmanager。
- 启动 Prometheus Alertmanager。
- 配置 Prometheus Alertmanager。更多信息，参考 [Prometheus 文档](https://www.prometheus.io/docs/alerting/latest/configuration/)。

OBAgent 提供默认的报警项，配置文件位于 `conf/prometheus_config/rules`。其中，`host_rules.yaml` 存储机器报警项，`ob_rules.yaml` 存储 OceanBase 数据库报警项。如果默认报警项不能满足您的需求，按照以下方式自定义报警项：

```yaml
#在 Prometheus 的配置文件中增加报警相关的配置。报警相关的配置文件需放在 rules 目录，且命名满足 *rule.yaml。

groups:
- name: node-alert
    rules:
    - alert: disk-full
    expr: 100 - ((node_filesystem_avail_bytes{mountpoint="/",fstype=~"ext4|xfs"} * 100) / node_filesystem_size_bytes {mountpoint="/",fstype=~"ext4|xfs"}) > 80
    for: 1m
    labels:
        serverity: page
    annotations:
        summary: "{{ $labels.instance }} disk full "
        description: "{{ $labels.instance }} disk > {{ $value }}  "
```

## （可选）更新 KV 配置

OBAgent 提供了更新配置的接口。您可以通过 HTTP 服务更新 KV 配置项：

```bash
# 您可以同时更新多个 KV 的配置项，写多组 key 和 value 即可。

curl --user user:pass -H "Content-Type:application/json" -d '{"configs":[{"key":"monagent.pipeline.ob.status", "value":"active"}]}' -L 'http://ip:port/api/v1/module/config/update'
```
