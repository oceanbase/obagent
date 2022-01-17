# 使用 OBD 部署 OBAgent

OBAgent 提供使用 ODB 部署和手动部署。要手动部署 OBAgent，您要配置 OBAgent、Prometheus 和 Prometheus Alertmanager（可选）。推荐您使用 OBD 部署 OBAgent。

## 前提条件

在部署 OBAgent 之前，您需要确认 OBAgent 的默认端口 8088、8089 未占用。您也可以自定义端口。

> **说明**：如果您的机器可以连接公网，在您执行了 `obd cluster deploy` 命令之后，OBD 将检查您的目标机器是否有 OBAgent 安装包。如果没有安装包，OBD 将自动从 yum 源获取。

## 同时部署 OceanBase 集群和 OBAgent

如果您希望同时部署 OceanBase 集群和 OBAgent，您只需要在 OceanBase 数据库的配置文件中添加以下 OBAgent 的配置信息：

```yaml
obagent:
  servers:
    - 127.0.0.1
  depends: 
    - oceanbase-ce
  global:
    home_path: /root/observer
```

> **注意**：`servers` 字段必须与 oceanbase-ce 的 `servers` 字段一致。

详细信息，请参考 [配置文件](https://github.com/oceanbase/obdeploy/blob/master/example/obagent/distributed-with-obproxy-and-obagent-example.yaml)。

## 单独部署 OBAgent

OBD 不支持为已部署的集群添加新的组件。如果您希望为要已部署的集群配置 OBAgent，您需要单独部署 OBAgent。若要单独部署 OBAgent，您需要准备 OBAgent 的配置文件。请确保配置文件中的以下字段与 OceanBase 数据库相同：

```yaml
obagent:
  global:
    # Username for HTTP authentication. The default value is admin.
    http_basic_auth_user: admin
    # Password for HTTP authentication. The default value is root.
    http_basic_auth_password: root
    # Username for debug service. The default value is admin.
    pprof_basic_auth_user: admin
    # Password for debug service. The default value is root.
    pprof_basic_auth_password: root
    
    # 以下配置必须与 OceanBase 数据库一致
    # Monitor username for OceanBase Database. The user must have read access to OceanBase Database as a system tenant. The default value is root.
    monitor_user: root
    # Monitor password for OceanBase Database. The default value is empty. When a depends exists, OBD gets this value from the oceanbase-ce of the depends. The value is the same as the root_password in oceanbase-ce.
    monitor_password: 
    # The SQL port for observer. The default value is 2881. When a depends exists, OBD gets this value from the oceanbase-ce of the depends. The value is the same as the mysql_port in oceanbase-ce.
    sql_port: 2881
    # The RPC port for observer. The default value is 2882. When a depends exists, OBD gets this value from the oceanbase-ce of the depends. The value is the same as the rpc_port in oceanbase-ce.
    rpc_port: 2882
    # Cluster name for OceanBase Database. When a depends exists, OBD gets this value from the oceanbase-ce of the depends. The value is the same as the appname in oceanbase-ce.
    cluster_name: obcluster
    # Cluster ID for OceanBase Database. When a depends exists, OBD gets this value from the oceanbase-ce of the depends. The value is the same as the cluster_id in oceanbase-ce.
    cluster_id: 1
    # Zone name for your observer. The default value is zone1. When a depends exists, OBD gets this value from the oceanbase-ce of the depends. The value is the same as the zone name in oceanbase-ce.
    zone_name: zone1
    # Monitor status for OceanBase Database.  Active is to enable. Inactive is to disable. The default value is active. When you deploy an cluster automatically, OBD decides whether to enable this parameter based on depends.
```

更多信息，请参考 [OBAgent 配置文件](https://github.com/oceanbase/obdeploy/blob/master/example/obagent/obagent-only-example.yaml
)。

## 启动 OBAgent

使用以下命令 启动 OBAgent：

```bash
# 传入配置信息
obd cluster deploy <deploy name> [-c <yaml path>] [-f] [-U] [-A]
# 启动 OBAgent
obd cluster start <deploy name> [flags]
```

更多信息，请参考 [OBD 使用文档](https://github.com/oceanbase/obdeploy/blob/master/README-CN.md#obd-cluster-deploy)。

## （可选）启动 Prometheus

> 说明：您需要安装 Prometheus。

OBD 会替换 OBAgent 配置文件模版中的变量，可以复制该配置文件（prometheus.yaml）到部署 Prometheus 的主机，然后执行如下命令使用 prometheus.yaml 文件启动 Prometheus：

```bash
./prometheus --config.file=./prometheus.yaml
```

## （可选）部署 Prometheus Alertmanager

- 下载并解压 Prometheus Alertmanager。
- 启动 Prometheus Alertmanager。
- 配置 Prometheus Alertmanager。更多信息，请参考 [Prometheus 文档](https://www.prometheus.io/docs/alerting/latest/configuration/)。

OBAgent 提供默认的报警项，配置文件位于 `conf/prometheus_config/rules`。其中，`host_rules.yaml` 存储机器报警项，`ob_rules.yaml` 存储 OceanBase 数据库报警项。如果默认报警项不能满足您的需求，按照以下方式自定义报警项：

```yaml
# 在 Prometheus 的配置文件中增加报警相关的配置。报警相关的配置文件需放在 rules 目录，且命名满足 *rule.yaml。

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

### （可选）部署与配置 Grafana

- 下载并解压 Grafana。
- 启动 Grafana。
    ```bash
    ./bin/grafana-server
    ```
- 登陆 grafana 并且修改密码，第一次登陆默认账号和密码都是 admin。
- 配置 prometheus 为数据源, 添加数据源的方式参考 [grafana 文档](https://grafana.com/docs/grafana/v7.5/datasources/add-a-data-source/)。
- 导入 dashboard, 针对 OB 和主机指标的配置已经发布到grafana，id 分别为 15215, 15216, 导入 dashboard 的方式请参考 [grafana 文档](https://grafana.com/docs/grafana/v7.5/dashboards/export-import/)。

## （可选）更新 KV 配置

要更新 KV 配置，请使用 `obd cluster edit-config`。详细信息，请参考 [OBD 使用文档](https://github.com/oceanbase/obdeploy/blob/master/README-CN.md#obd-cluster-edit-config)。
