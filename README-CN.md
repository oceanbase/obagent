# OBAgent

OBAgent 是一个监控采集框架。OBAgent 支持推、拉两种数据采集模式，可以满足不同的应用场景。OBAgent 默认支持的插件包括主机数据采集、OceanBase 数据库指标的采集、监控数据标签处理和 Prometheus 协议的 HTTP 服务。要使 OBAgent 支持其他数据源的采集，或者自定义数据的处理流程，您只需要开发对应的插件即可。

## 许可证

OBAgent 使用 [MulanPSL - 2.0](http://license.coscl.org.cn/MulanPSL2) 许可证。您可以免费复制及使用源代码。当您修改或分发源代码时，请遵守木兰协议。

## 文档

参考 [OBAgent 文档](docs/about-obagent/what-is-obagent.md)。

## 如何获取

### 环境依赖

构建 OBAgent 需要 Go 1.14 版本及以上。

### RPM 包

OBAgent 提供 RPM 包，您可以去 [Release 页面](https://mirrors.aliyun.com/oceanbase/community/stable/el/7/x86_64/) 下载 RPM 包，然后使用以下命令安装：

```bash
rpm -ivh obagent-0.1-1.alios7.x86_64.rpm
```

### 通过源码构建

#### Debug 模式

```bash
make build // make build will be debug mode by default
make build-debug
```

#### Release 模式

```bash
make build-release
```

## 如何开发

您可以为 OBAgent 开发插件。更多信息，参考 [OBAgent 插件开发](docs/develop-guide.md)。

## 如何贡献

我们十分欢迎并感谢您为我们贡献。以下是您参与贡献的几种方式：

- 向我们提 [Issue](https://github.com/oceanbase/obagent/issues)。

## 获取帮助

如果您在使用 OBAgent 时遇到任何问题，欢迎通过以下方式寻求帮助：

- [GitHub Issue](https://github.com/oceanbase/obagent/issues)
- [官方网站](https://open.oceanbase.com/)
- 知识问答（Coming soon）
