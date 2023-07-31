# 什么是 OBAgent

OBAgent 是一个监控采集和运维框架。
监控部分，OBAgent 支持推、拉两种数据采集模式，可以满足不同的应用场景。OBAgent 默认支持的插件包括主机数据采集、OceanBase 数据库指标的采集、日志信息采集、监控数据标签处理和 Prometheus 协议的 HTTP 服务。支持推送数据到pushgateway、vmagent、es、sls、alertmanager。要使 OBAgent 支持其他数据源的采集，或者自定义数据的处理流程，您只需要开发对应的插件即可。
运维部分，OBAagent 支持配置热更新，同时提供了对agent自身运维的接口，以及文件操作和rpm包操作的接口，您可以根据自身需求来使用这些接口。

## 特性

OBAgent 具有以下特性：

- 采用 Go 语言开发，无需外部依赖。
- 插件驱动，易开发。

# OBAgent 架构

OBAgent 由四部分组成，包含 agentd、mgragent、monagent 以及 agentctl。

进程名称 | 说明
--- | ---
agentd | 作为守护进程，用于启停 mgragent & monagent 进程，如果 mgragent 或者 monagent 非正常退出，agentd 会负责将他们自动重新启动起来，保证服务可用；同时 agentd 也会运行一些资源限制任务，用于限制 mgragent / monagent 的资源使用情况（比如 CPU、内存等）。
mgragent | 提供运维相关的功能。包括了重启、重装OBAgent，文件管理，安装包管理，以及配置管理等。
monagent | 提供了监控功能。监控的内容范围包含了数据库性能参数、应用日志、主机性能/资源参数、租户资源/性能参数、SQL/事务参数等等。用户可以通过 推 和 拉 两种模式获取监控采集的数据内容。
agentctl | 提供了黑屏运维功能。agentctl 本身不包含业务逻辑，以命令行的方式提供给用户一个操作入口。包含了进程启停，配置变更等。


