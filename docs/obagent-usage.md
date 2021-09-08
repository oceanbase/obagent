# 安装和配置
obagent提供rpm包，可以使用rpm命令进行安装
// TODO: replace with real url
[obagent-0.1-1.alios7.x86_64.rpm]()

```bash
rpm -ivh obagent-0.1-1.alios7.x86_64.rpm
```
​

**目录结构**
安装之后会有一个binary文件，和一组配置文件，配置文件中又分为程序启动配置，模块配置模版和kv变量的配置, 另外为了方便使用还有一个prometheus的配置模版
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
​

**monagent.yaml配置文件:**
```bash
## 日志相关配置
log:
  level: debug
  filename: log/monagent.log
  maxsize: 30
  maxage: 7
  maxbackups: 10
  localtime: true
  compress: true

## 进程相关配置，address是默认的拉取metric和管理相关接口，adminAddress是pprof调试端口
server:
  address: "0.0.0.0:8088"
  adminAddress: "0.0.0.0:8089"
  runDir: run

## 配置相关，加密方法支持aes和plain，如果是aes，会使用下面key文件中的key对需要加密的配置项进行加密
## modulePath中存放配置模版，propertiesPath存放kv变量配置
cryptoMethod: plain
cryptoPath: conf/.config_secret.key
modulePath: conf/module_config
propertiesPath: conf/config_properties
```
​

**配置模版**
```bash
## basic auth 相关配置，可以配置两个端口开启或者关闭，配置disabled后对应的变量 {disable_http_basic_auth} {disable_pprof_basic_auth}
monagent_basic_auth.yaml 

## 配置模块相关的配置，一般不需要修改
monagent_config.yaml

## 主机监控流水线配置模版,一般不需要修改
monitor_node_host.yaml

## OB监控流水线配置模版，一般不需要修改
monitor_ob.yaml
```


**kv配置项**
```bash
## basic auth 相关的kv配置项
monagent_basic_auth.yaml

## 流水线相关的kv配置项
monagent_pipeline.yaml
```
**​**

**kv配置项说明:**
```yaml
# encrypted=true的配置项, 需要加密存储，目前支持aes加密方法，
# {}中的变量需要进行替换，替换成真实的值，如果monagent启动配置中设置了加密方法=aes, 需要配置加密之后的值

## basic auth
configVersion: "2021-08-20T07:52:28.5443+08:00"
configs:
    - key: http.server.basic.auth.username
      value: {http_basic_auth_user}
      valueType: string
    - key: http.server.basic.auth.password
      value: {http_basic_auth_password}
      valueType: string
      encrypted: true
    - key: http.admin.basic.auth.username
      value: {pprof_basic_auth_user}
      valueType: string
    - key: http.admin.basic.auth.password
      value: {pprof_basic_auth_password}
      valueType: string
      encrypted: true
      
## pipeline
configVersion: "2021-08-20T07:52:28.5443+08:00"
configs:
    - key: monagent.ob.monitor.user
      value: {monitor_user}
      valueType: string
    - key: monagent.ob.monitor.password
      value: {monitor_password}
      valueType: string
      encrypted: true
    - key: monagent.ob.sql.port
      value: {sql_port}
      valueType: int64
    - key: monagent.ob.rpc.port
      value: {rpc_port}
      valueType: int64
    - key: monagent.host.ip
      value: {host_ip}
      valueType: string
    - key: monagent.ob.cluster.name
      value: {cluster_name}
      valueType: string
    - key: monagent.ob.cluster.id
      value: {cluster_id}
      valueType: int64
    - key: monagent.ob.zone.name
      value: {zone_name}
      valueType: string
    - key: monagent.pipeline.ob.status
      value: {ob_monitor_status}
      valueType: string
    - key: monagent.pipeline.node.status
      value: {host_monitor_status}
      valueType: string

```
**​**

**启动monagent:**
```bash
# 推荐使用supervisor来拉起进程
# 启动命令
cd /home/admin/obagent
nohup ./bin/monagent -c conf/monagent.yaml >> ./log/monagent_stdout.log 2>&1 &

# supervisor 配置样例
[program:monagent]
command=./bin/monagent -c conf/monagent.yaml
directory=/home/admin/obagent
autostart=true
autorestart=true
redirect_stderr=true
priority=10
stdout_logfile=log/monagent_stdout.log
```
**​**

**配置更新:**
obagent提供了配置更新的接口, 可以通过http服务的方式更新kv配置项，具体的调用方式
```bash
# 可以同时更新多个kv的配置项，写多组key和value的值即可

curl --user user:pass -H "Content-Type:application/json" -d '{"configs":[{"key":"monagent.pipeline.ob.status", "value":"active"}]}' -L 'http://ip:port/api/v1/module/config/update'
```
# prometheus采集配置
**prometheus配置样例**
```yaml
# obagent的rpm包中携带了一份prometheus的配置模版，可以根据实际情况做一些修改，
# 如果开启basic auth认证，需要配置{http_basic_auth_user} {http_basic_auth_password}
# {target} 替换成主机的ip + port
# rules 目录下有两个报警配置模版，分别是默认的主机和ob报警配置，如需自定义报警项，可以作为参考

global:
  scrape_interval:     1s
  evaluation_interval: 10s

rule_files:
  - "rules/*rules.yaml"

scrape_configs:
  - job_name: prometheus
    metrics_path: /metrics
    scheme: http
    static_configs:
    - targets:
      - 'localhost:9090'
  - job_name: node
    basic_auth:
      username: {http_basic_auth_user}
      password: {http_basic_auth_password}
    metrics_path: /metrics/node/host
    scheme: http
    static_configs:
      - targets:
        - {target}
  - job_name: ob_basic
    basic_auth:
      username: {http_basic_auth_user}
      password: {http_basic_auth_password}
    metrics_path: /metrics/ob/basic
    scheme: http
    static_configs:
      - targets:
        - {target}
  - job_name: ob_extra
    basic_auth:
      username: {http_basic_auth_user}
      password: {http_basic_auth_password}
    metrics_path: /metrics/ob/extra
    scheme: http
    static_configs:
      - targets:
        - {target}
  - job_name: agent
    basic_auth:
      username: {http_basic_auth_user}
      password: {http_basic_auth_password}
    metrics_path: /metrics/stat
    scheme: http
    static_configs:
      - targets:
        - {target}
```


**启动prometheus**
```yaml
# 需要首先下载prometheus

./prometheus --config.file=./prometheus.yaml
```
**​**

**在prometheus中查看exporter状态**
// TODO: replace with real url
![Screenshot 2021-08-12 at 20.29.35.png](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/28412/1628771389150-58415f6b-f455-416a-8329-46822eb292dc.png#clientId=u9961547e-7621-4&from=ui&id=u8d70909f&margin=%5Bobject%20Object%5D&name=Screenshot%202021-08-12%20at%2020.29.35.png&originHeight=964&originWidth=1980&originalType=binary&ratio=1&size=198565&status=done&style=none&taskId=u8a79e4af-ee50-4d8b-bd92-3ae6f780cdd)


**在prometheus中计算监控指标**
![Screenshot 2021-08-12 at 20.32.53.png](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/28412/1628771585474-6b14d363-22d5-4748-a032-299c5a08484c.png#clientId=u9961547e-7621-4&from=ui&id=u8297d38a&margin=%5Bobject%20Object%5D&name=Screenshot%202021-08-12%20at%2020.32.53.png&originHeight=842&originWidth=2558&originalType=binary&ratio=1&size=177625&status=done&style=none&taskId=uf184d0a6-af77-4418-84d2-2b3fb736d00)


**配置报警相关信息:**

1. **部署alertmanager**
```yaml
1. 下载alertmanager
2. 解压
3. 启动
./alertmanager --config.file=alertmanager.yaml

具体的配置信息可以参考 https://www.prometheus.io/docs/alerting/latest/configuration/
```

2. **配置prometheus**
```yaml
# prometheus的配置文件中增加报警相关的配置, 根据上面的配置文件，报警相关的配置文件放在rules目录下，命名满足*rule.yaml
以磁盘监控为例

groups:
  - name: node-alert
    rules:
    - alert: disk-full
      expr: 100 - ((node_filesystem_avail_bytes{mountpoint="/",fstype=~"ext4|xfs"} * 100) / node_filesystem_size_bytes {mountpoint="/",fstype=~"ext4|xfs"}) > 80
      for: 1m
      labels:
        serverity: page
      annotations:
        summary: "{{ $labels.instance }} disk full "
        description: "{{ $labels.instance }} disk > {{ $value }}  "

```

3. **查看报警信息**

![Screenshot 2021-09-05 at 22.16.57.png](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/28412/1630851428176-e5dd994c-b6ad-4e56-8cf5-a93dda31d639.png#clientId=udbbe738a-dc94-4&from=ui&id=uca310bcd&margin=%5Bobject%20Object%5D&name=Screenshot%202021-09-05%20at%2022.16.57.png&originHeight=1326&originWidth=2880&originalType=binary&ratio=1&size=337214&status=done&style=none&taskId=u9ae5bd9d-3c54-4581-9ea8-dae5f49e898)
**​**

