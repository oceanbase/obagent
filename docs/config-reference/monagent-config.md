# monagent 配置文件说明

本文介绍 monagent 配置文件中的相关配置项，并列出了配置文件模板供您参考。

`monagent.yaml` 配置文件的示例如下：

```yaml
## 日志相关配置
log:
  level: debug
  filename: log/monagent.log
  maxsize: 30
  maxage: 7
  maxbackups: 10
  localtime: true
  compress: true

## 进程相关配置。其中，address 是默认的拉取 metrics 和管理相关接口，adminAddress 是 pprof 调试端口。
server:
  address: "0.0.0.0:8088"
  adminAddress: "0.0.0.0:8089"
  runDir: run

## 配置相关，加密方法支持 aes 和 plain。其中，aes 使用下面 key 文件中的 key 对需要加密的配置项进行加密。
## modulePath 用来存放配置模版，propertiesPath 用来存放 KV 变量配置
cryptoMethod: plain
cryptoPath: conf/.config_secret.key
modulePath: conf/module_config
propertiesPath: conf/config_properties
```

## 配置模版

monagent 的相关配置文件模板见下表：

配置文件名称 | 说明
--- | ---
monagent_basic_auth.yaml | 基础认证的相关配置，用来配置两个端口的开启或者关闭，配置禁用后对应的变量为 {disable_http_basic_auth}  和 {disable_pprof_basic_auth}。
monagent_config | 配置模块相关的配置，无需修改。
monitor_node_host.yaml | 主机监控流水线配置模版，无需修改。
monitor_ob.yaml | OceanBase 数据库监控流水线配置模版，无需修改。
