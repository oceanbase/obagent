# OB-Agent
OB Agent 是 OCP Express 远程访问主机和 OBServer 的入口，提供两个主要能力：
1. 运维主机和 OBServer；
2. 收集主机、OBServer的监控。 
此外，OB Agent 还承担 OB 日志查询和日志清理的职能。

## 目录结构

```
obagent
├── cmd：程序主入口
│   ├── mgragent: 运维进程主入口，后台运行的进程，不期望被用户使用；目标产物名称为 ob_mgragent。
│   ├── monagent: 监控进程主入口，后台运行进程，不期望被用户使用；目标产物名称为 ob_monagent。
│   ├── agentd: 守护进程，后台运行的进程，负责拉起异常退出的ob_mgragent和ob_monagent进程；目标产物名称为ob_agentd。
│   └── agentctl: 黑屏运维工具主入口，命令行运维工具，可通过该工具运维 ob_mgragent 和 ob_monagent 进程；目标产物名称为 ob_agentctl 。
├── api：请求处理、认证
│   ├── monroute：监控模块route和handlers
│   ├── mgrroute：运维模块route和handlers
│   ├── web：http server
│   └── common：公共handlers和middleware
├── agentd：守护进程agentd模块
│   └── api：进程状态信息
├── config: 配置文件解析，回调函数注册
│   ├── monconfig：监控配置
│   ├── mgrconfig：运维配置
│   ├── agentctl：黑屏运维工具配置
│   └── sdk：配置管理，回调函数注册
├── executor：提供运维能力，命令执行能力，支持 shell 跨平台执行
│   ├── agent：agent进程运维
│   ├── cleaner：日志清理
│   ├── log_query：日志查询
│   └── ...
├── lib：通用方法，比如加密、脱敏、重试、命令行执行等
│   ├── command：命令执行，异步任务调度
│   ├── process：进程启停
│   ├── goroutinepool：任务池
│   ├── log_analyzer：日志解析
│   ├── retry：重试框架
│   ├── shellf：命令模板配置解析，命令构建
│   ├── shell：命令执行
│   └── ...
├── monitor：监控模块，包含插件定义、流水线引擎、监控数据结构等
│   ├── engine：流水线引擎
│   ├── plugins：流水线插件
│   ├── message：监控数据结构
│   └── utils：监控通用函数
├── stat：自监控模块，obagent、moniotr 都会依赖此模块实现自监控
├── log：日志框架
├── errors：错误处理
├── rpm：rpm打包逻辑
├── tests：测试脚本、数据、配置
├── etc：发布的配置文件，均为 yaml 类型
│   ├── config_properties：key-value配置
│   ├── module_config：流水线等配置文件
│   ├── prometheus_config：prometheus拉取配置
│   ├── scripts：开机自启动脚本
│   └── shell_templates：命令模板
└── docs：文档，包括 obagent 的 README 文档，以及各个子模块的说明文档
```

# 安装 OBAgent

您可以使用 RPM 包或者构建源码安装 OBAgent。

## 环境依赖

构建 OBAgent 需要 Go 1.19 版本及以上。

## RPM 包

OBAgent 提供 RPM 包，您可以去 [Release 页面](https://mirrors.aliyun.com/oceanbase/community/stable/el/7/x86_64/) 下载 RPM 包，然后使用以下命令安装：

```bash
rpm -ivh obagent-4.1.0-1.el7.x86_64.rpm
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

# License
OBAgent根据 Mulan 公共许可证版本 2 获得许可。有关详细信息，请参阅 [木兰宽松许可证, 第2版](http://license.coscl.org.cn/MulanPSL2) 。
