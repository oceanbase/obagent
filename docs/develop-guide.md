# OBAgent 开发指南

OBAgent 是一个插件驱动的监控采集框架。要扩展 OBAgent 的功能，或者自定义数据的处理流程，您可以开发对应的插件。开发插件时，您只需要实现插件的基本接口和对应类型插件的接口即可。

## OBAgent 数据处理流程

![OBAgent数据处理流程图](https://github.com/Xjxjy/obagent/blob/master/picture/OBAgent-Process.png)

OBAgent 的数据处理流程包括数据采集、处理和上报，需要用到的插件包含输入插件（Source）、处理插件（Processor）、输出插件（Sink）。插件详细信息，参考 [外部插件](#外部插件) 章节。

## 外部插件

OBAgent 支持的插件类型见下表：

| 插件类型            | 功能描述                                       |
|-----------------|--------------------------------------------|
| 输入插件（Source）    | 收集各种时间序列性指标，包含各种系统信息和应用信息的插件。              |
| 处理插件（Processor） | 串行进行数据处理。                                  |
| 输出插件（Sink）      | 数据输出，支持推和拉两种模式。                            |

### 输入插件接口定义

```go
type Source interface {
    Start(out chan<- []*message.Message) (err error)
    Stop()
}
```

输入插件在Start中开始采集数据。并将数据写入out中。
​

### 处理插件接口定义

```go
type Processor interface {
    Start(in <-chan []*message.Message, out chan<- []*message.Message) (err error)
    Stop()
}
```

处理插件从in中读数据，处理后写到out中。
​

### 输出插件接口定义

```go
type Sink interface {
    Start(in <-chan []*message.Message) error
    Stop()
}
```

输出插件从in中读取数据，如果是拉模式，则放入缓存中等待拉取；如果是推模式，则直接推给目的端。

## message 数据结构

OBAgent 数据处理流程中流转的数据定义为统一的 message。

```go
type Message struct {
    name        string
    fields      []FieldEntry
    tags        []TagEntry
    timestamp   time.Time
    msgType     Type
    tagSorted   bool
    fieldSorted bool
    id          string
}
```
