# monagent 配置文件说明

本文介绍 monagent.yaml 配置文件的相关配置项，并列出了配置文件模板供您参考。这里面的大部分配置项都是用了${config_key}来进行表示，${config_key}和config_properties下KV-config关联，并在启动agent时进行替换。

`monagent.yaml` 配置文件的示例如下：

```yaml

## 进程相关配置。其中，address 是默认的拉取 metrics 和管理相关接口，也是pprof调试端口。
server:
  address: 0.0.0.0:${ocp.agent.monitor.http.port}
  runDir: ${obagent.home.path}/run

## 配置相关，加密方法支持 aes 和 plain。其中，aes 使用下面 key 文件中的 key 对需要加密的配置项进行加密。
## modulePath 用来存放配置模版，propertiesPath 用来存放 KV 变量配置
cryptoMethod: aes
cryptoPath: ${obagent.home.path}/conf/.config_secret.key
modulePath: ${obagent.home.path}/conf/module_config
propertiesPath: ${obagent.home.path}/conf/config_properties
```

