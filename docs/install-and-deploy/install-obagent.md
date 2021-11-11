# 安装 OBAgent

您可以使用 RPM 包或者构建源码安装 OBAgent。

## 环境依赖

构建 OBAgent 需要 Go 1.14 版本及以上。

## RPM 包

OBAgent 提供 RPM 包，您可以去 [Release 页面](https://mirrors.aliyun.com/oceanbase/community/stable/el/7/x86_64/) 下载 RPM 包，然后使用以下命令安装：

```bash
rpm -ivh obagent-1.0.0-1.el7.x86_64.rpm
```

## 构建源码

### Debug 模式

```bash
make build // make build will be debug mode by default
make build-debug
```

### Release 模式

```bash
make build-release
```

## OBAgent 安装目录结构

OBAgent 的安装目录包含三个子目录：`bin`、`conf` 和 `run`。OBAgent 的安装目录如下：

```bash
# 目录结构示例
.
├── bin
│   └── monagent
├── conf
│   ├── config_properties
│   │   ├── monagent_basic_auth.yaml
│   │   └── monagent_pipeline.yaml
│   ├── module_config
│   │   ├── monagent_basic_auth.yaml
│   │   ├── monagent_config.yaml
│   │   ├── monitor_node_host.yaml
│   │   └── monitor_ob.yaml
│   ├── monagent.yaml
│   └── prometheus_config
│       ├── prometheus.yaml
│       └── rules
│           ├── host_rules.yaml
│           └── ob_rules.yaml
└── run
```

其中，`bin` 用来存放二进制文件。`conf` 用来存放程序启动配置、模块配置模板、KV 变量配置和 Prometheus 的配置模板。`run` 用来存放运行文件。更多关于配置文件的信息，参考 [monagent 配置文件](../config-reference/monagent-config.md)。
