# 压力测试
压力测试目标：TODO

## 配置文件

TODO

## 指标数据

1. 配置grafana 看板
2. 启动prometheus
```
prometheus --config.file prometheus.yaml
```
3. sysbench 基准测试

注意： ob 默认关闭PS，需要临时打开
```
alter system set  _ob_enable_prepared_statement = True
```
