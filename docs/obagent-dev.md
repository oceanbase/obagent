# obagent数据处理流程
![Screenshot 2021-09-15 at 11.36.11.png](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/28412/1631676986085-49f40134-9502-438b-bb32-5a3ee6591fbc.png#clientId=u89a16740-3189-4&from=ui&id=ucf4512dc&margin=%5Bobject%20Object%5D&name=Screenshot%202021-09-15%20at%2011.36.11.png&originHeight=868&originWidth=1638&originalType=binary&ratio=1&size=108483&status=done&style=none&taskId=uc6b3bf07-12ab-4e56-9568-bdab0001cc4)
obagent将插件组合，定义成流水线，流水线作为基本的调度单位，实现一个完整的数据采集，处理和上报流程，同时支持两种模式，推模式和拉模式，一个流水线包含一组input插件，并行进行数据采集，一组processor插件，串行进行数据处理，拉模式包括一个exporter插件，通过http服务的方式将数据暴露出来，推模式包括一个output插件，实现数据推送功能,  流水线中流转的数据定义为统一的Metric接口，为了扩展obagent的能力，可以自定义开发一些插件，插件的开发只需要实现插件的基本接口和对应类型插件的接口即可
​

# metric接口定义
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


# 插件基本接口定义
```go
//Initializer contains Init function
type Initializer interface {
	Init(config map[string]interface{}) error
}

//Closer contains Close function
type Closer interface {
	Close() error
}

//Describer show SampleConfig and Description
type Describer interface {
	SampleConfig() string
	Description() string
}

```
所有obagent的插件都必须实现以上的基本接口

- Init方法用来做插件的初始化工作
- Close方法在插件退出时调用，用来关闭一些资源
- SampleConfig方法返回插件的配置样例
- Description方法返回插件的描述信息

​

# 不同类型插件的接口定义
## Input插件
```go
type Input interface {
	Collect() ([]metric.Metric, error)
}
```
intput 插件需要实现Collect方法, 来采集数据，返回一组metric和error
​

## Processor插件
```go
type Processor interface {
    Process(metrics ...metric.Metric) ([]metric.Metric, error)
}
```
processor插件需要实现Process方法，用来做数据处理，输入参数是一组metric，输出一组metric和error
​

## Exporter插件
```go
type Exporter interface {
	Export(metrics []metric.Metric) (*bytes.Buffer, error)
}
```
exporter插件需要实现Export方法，输入参数是一组metric，输出byte buffer 和 error， 作用是将metric做格式转换，只用在拉模式
​

## Output插件
```go
type Output interface {
	Write(metrics []metric.Metric) error
}
```
output插件需要实现Write方法，输入参数是一组metric，输出error，作用是将metric数据推送到远端，只用在推模式
