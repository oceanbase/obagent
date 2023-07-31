# mgragent 配置文件说明

本文介绍 mgragent.yaml 配置文件的相关配置项，并列出了配置文件模板供您参考。这里面的大部分配置项都是用了${config_key}来进行表示，${config_key}和config_properties下KV-config关联，并在启动agent时进行替换。

`monagent.yaml` 配置文件的示例如下：

```yaml

## 安装相关配置。指定 obagent 的home路径。
install:
  path: ${obagent.home.path}

## 进程相关配置。其中，address 是默认的拉取 metrics 和管理相关接口，也是pprof调试端口。
server:
  address: 0.0.0.0:${ocp.agent.monitor.http.port}
  runDir: ${obagent.home.path}/run

## sdk 配置相关，加密方法支持 aes 和 plain。其中，aes 使用下面 key 文件中的 key 对需要加密的配置项进行加密。
## moduleConfigDir 用来存放配置模版，configPropertiesDir 用来存放 KV 变量配置
sdkConfig:
  configPropertiesDir: ${obagent.home.path}/conf/config_properties
  moduleConfigDir: ${obagent.home.path}/conf/module_config
  cryptoPath: ${obagent.home.path}/conf/.config_secret.key
  cryptoMethod: aes

## 命令模板配置相关。指定mgragent的配置模板文件。
shellf:
  template: ${obagent.home.path}/conf/shell_templates/shell_template.yaml
```
