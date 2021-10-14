# OBAgent 开发指南

OBAgent 是一个插件驱动的监控采集框架。要扩展 OBAgent 的功能，或者自定义数据的处理流程，您可以开发对应的插件。开发插件时，您只需要实现插件的基本接口和对应类型插件的接口即可。

## OBAgent 数据处理流程

![Screenshot 2021-09-15 at 11.36.11.png](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/28412/1631676986085-49f40134-9502-438b-bb32-5a3ee6591fbc.png#clientId=u89a16740-3189-4&from=ui&id=ucf4512dc&margin=%5Bobject%20Object%5D&name=Screenshot%202021-09-15%20at%2011.36.11.png&originHeight=868&originWidth=1638&originalType=binary&ratio=1&size=108483&status=done&style=none&taskId=uc6b3bf07-12ab-4e56-9568-bdab0001cc4)

OBAgent 的数据处理流程包括数据采集、处理和上报，需要用到的插件包含输入插件（Inputs）、处理插件（Process）、输出插件（OutPuts 和 Exporter）。插件详细信息，参考 [外部插件](#外部插件) 章节。

## 外部插件

OBAgent 支持的插件类型见下表：

| 插件类型              | 功能描述                                                                                                                                                                   |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 输入插件（Input）    | 收集各种时间序列性指标，包含各种系统信息和应用信息的插件。                                                                                                                 |
| 处理插件（Process）   | 串行进行数据处理。                                                              |
| 输出插件（Output）   | 仅适用于推模式。用来将 metrics 数据推送到远端。     |
输出插件（Exporter） | 仅适用于拉模式。通过 HTTP 服务暴露数据。用来为 metrics 做格式转换。

### 输入插件接口定义

```go
type Input interface {
    Collect() ([]metric.Metric, error)
}
```

输入插件需要实现 Collect 方法采集数据。输入插件返回一组 metrics 和 error。
​

### 处理插件接口定义

```go
type Processor interface {
    Process(metrics ...metric.Metric) ([]metric.Metric, error)
}
```

处理插件需要实现 Process 方法处理数据，输入参数是一组 metrics，输出一组 metrics 和 error。
​

### 输出插件（Exporter）接口定义

```go
type Exporter interface {
    Export(metrics []metric.Metric) (*bytes.Buffer, error)
}
```

输出插件（Exporter）需要实现 Export 方法，输入参数是一组 metrcs，输出 byte buffer 和 error。

### 输出插件（Output）接口定义

```go
type Output interface {
    Write(metrics []metric.Metric) error
}
```

输出插件（Output）需要实现 Write 方法，输入参数是一组 metrics，输出 error。

## Metric 接口定义

OBAgent 数据处理流程中流转的数据定义为统一的 Metric 接口。

```go
type Metric interface {
        Clone() Metric
        SetName(name string)
        GetName() string
        SetTime(time time.Time)
        GetTime() time.Time
        SetMetricType(metricType Type)
        GetMetricType() Type
        Fields() map[string]interface{}
        Tags() map[string]string
}
```

## 插件基本接口定义

所有的 OBAgent 插件都必须实现以下的基本接口：

```go
//Initializer 包含 Init 函数
type Initializer interface {
    Init(config map[string]interface{}) error
}

//Closer 包含 Close 函数
type Closer interface {
    Close() error
}

//Describer 包含 SampleConfig 和 Description
type Describer interface {
    SampleConfig() string
    Description() string
}

```

函数详情见下表：

函数名 | 说明
--- | ---
Init | 初始化插件。
Close | 在插件退出时调用，用来关闭一些资源。
SampleConfig | 用来返回插件的配置样例。
Description | 用来返回插件的描述信息。
