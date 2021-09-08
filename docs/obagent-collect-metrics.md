# 常用的监控指标
**exporter暴露的指标**

| **类别** | **指标名** | **label** | **含义** | **类型** |
| --- | --- | --- | --- | --- |
| **主机**
​
 | node_xxx | node_exporter的label再增加ob_cluster_id,ob_cluster_name,obzone,svr_ip | 主机监控指标 | 参考node_exporter对应的指标类型 |
| **OB** | ob_active_session_num | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | 活跃连接数 | gauge |
|  | ob_cache_size_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name,cache_name | kvcache大小 | gauge |
|  | ob_partition_num | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | 分区数 | gauge |
|  | ob_plan_cache_access_total | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | 执行计划访问次数 | counter |
|  | ob_plan_cache_hit_total | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | 执行计划命中次数 | counter |
|  | ob_plan_cache_memory_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | plancache大小 | gauge |
|  | ob_table_num | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | 表数量 | gauge |
|  | ob_waitevent_wait_seconds_total | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | 等待事件总等待时间 | counter |
|  | ob_waitevent_wait_total | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | 等待事件总等待次数 | counter |
|  | ob_disk_free_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip | OB磁盘剩余大小 | gauge |
|  | ob_disk_total_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip | OB磁盘总大小 | gauge |
|  | ob_memstore_active_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | 活跃memstore大小 | gauge |
|  | ob_memstore_freeze_times | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | memstore冻结次数 | counter |
|  | ob_memstore_freeze_trigger_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | memstore冻结阈值 | gauge |
|  | ob_memstore_total_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name | memstore总大小 | gauge |
|  | ob_server_resource_cpu | ob_cluster_id,ob_cluster_name,obzone,svr_ip | observer可用cpu数 | gauge |
|  | ob_server_resource_cpu_assigned | ob_cluster_id,ob_cluster_name,obzone,svr_ip | observer已分配cpu数 | gauge |
|  | ob_server_resource_memory_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip | observer可用内存大小 | gauge |
|  | ob_server_resource_memory_assigned_bytes | ob_cluster_id,ob_cluster_name,obzone,svr_ip | observer已分配内存大小 | gauge |
|  | ob_unit_num | ob_cluster_id,ob_cluster_name,obzone,svr_ip | observer unit数量 | gauge |
|  | ob_sysstat | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name,stat_id | ob内部统计项 | 不同stat_id不相同，参考对应部分解释 |

**sysstat统计项**

| **stat_id** | **含义** | **类型** |
| --- | --- | --- |
| 10000 | 收到的rpc包数量 | counter |
| 10001 | 收到的rpc包大小 | counter |
| 10002 | 发送的rpc包数量 | counter |
| 10003 | 发送的rpc包大小 | counter |
| 10005 | rpc网络延迟 | counter |
| 10006 | rpc框架延迟 | counter |
| 20001 | 请求出队列次数 | counter |
| 20002 | 请求在队列中时间 | counter |
| 30000 | clog 同步时间 | counter |
| 30001 | clog 同步次数 | counter |
| 30002 | clog 提交次数 | counter |
| 30005 | 事务数 | counter |
| 30006 | 事务总时间 | counter |
| 40000 | select sql数 | counter |
| 40001 | select sql执行时间 | counter |
| 40002 | insert sql数 | counter |
| 40003 | insert sql执行时间 | counter |
| 40004 | replace sql数 | counter |
| 40005 | replace sql执行时间 | counter |
| 40006 | update sql数 | counter |
| 40007 | update sql执行时间 | counter |
| 40008 | delete sql数 | counter |
| 40009 | delete sql执行时间 | counter |
| 40010 | 本地sql执行次数 | counter |
| 40011 | 远程sql执行次数 | counter |
| 40012 | 分布式sql执行次数 | counter |
| 50000 | row cache命中次数 | counter |
| 50001 | row cache没有命中次数 | counter |
| 50008 | block cache命中次数 | counter |
| 50009 | block cache没有命中次数 | counter |
| 60000 | io 读次数 | counter |
| 60001 | io 读延时 | counter |
| 60002 | io 读字节数 | counter |
| 60003 | io 写次数 | counter |
| 60004 | io 写延时 | counter |
| 60005 | io 写字节数 | counter |
| 60019 | memstore读锁成功次数 | counter |
| 60020 | memstore读锁失败次数 | counter |
| 60021 | memstore写锁成功次数 | counter |
| 60022 | memstore写锁成功次数 | counter |
| 60023 | memstore等写锁时间 | counter |
| 60024 | memstore等读锁时间 | counter |
| 80040 | clog写次数 | counter |
| 80041 | clog写时间 | counter |
| 80057 | clog大小 | counter |
| 130000 | 活跃memstore大小 | gauge |
| 130001 | memstore总大小 | gauge |
| 130002 | 触发major freeze阈值 | gauge |
| 130004 | memstore大小限制 | gauge |
| 140002 | 最大可使用内存 | gauge |
| 140003 | 已使用内存 | gauge |
| 140005 | 最大可使用cpu | gauge |
| 140006 | 已使用cpu | gauge |

**​**

**常用指标的查询表达式**
实际查询的时候需要将变量替换成需要查询的具体信息

- @LABELS 替换为具体label的过滤条件
- @INTERVAL 替换为计算周期
- @GBLABELS 替换为聚合的label名称

| **指标** | **表达式** | **单位** |
| --- | --- | --- |
| 活跃 MEMStore 大小 | sum(ob_sysstat{stat_id="130000",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| 当前活跃会话数 | sum(ob_active_session_num{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
  |  |
| 块缓存命中率 | 100 * 1 / (1 + sum(rate(ob_sysstat{stat_id="50009",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="50008",@LABELS}[@INTERVAL])) by (@GBLABELS))  | % |
| 块缓存大小 | sum(ob_cache_size_bytes{cache_name="user_block_cache",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| 每秒提交的事务日志大小 | sum(rate(ob_sysstat{stat_id="80057",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | byte |
| 每次事务日志写盘平均耗时 | sum(rate(ob_sysstat{stat_id="80041",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="80040",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| CPU 使用率 | 100 * (1 - sum(rate(node_cpu_seconds_total{mode="idle", @LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(node_cpu_seconds_total{@LABELS}[@INTERVAL])) by (@GBLABELS))  | % |
| 磁盘分区已使用容量 | sum(node_filesystem_size_bytes{@LABELS} - node_filesystem_avail_bytes{@LABELS}) by (@GBLABELS) 
 / 1073741824  | GB |
| 每秒读次数 | avg(rate(node_disk_reads_completed_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | 次/s |
| 每次读取数据量 | avg(rate(node_disk_read_bytes_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| SSStore 每秒读次数 | sum(rate(ob_sysstat{stat_id="60000",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| SSStore 每次读取平均耗时 | sum(rate(ob_sysstat{stat_id="60001",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="60000",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| SSStore 每秒读取数据量 | sum(rate(ob_sysstat{stat_id="60002",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | byte |
| 每秒读取平均耗时 | 1000000 * (avg(rate(node_disk_read_time_seconds_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) / (avg(rate(node_disk_reads_completed_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) | us |
| 平均每次 IO 读取耗时 | 1000000 * (avg(rate(node_disk_read_time_seconds_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) / (avg(rate(node_disk_reads_completed_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) | us |
| 每秒写次数 | avg(rate(node_disk_writes_completed_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | 次/s |
| 每次写入数据量 | avg(rate(node_disk_written_bytes_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| SSStore 每秒写次数 | sum(rate(ob_sysstat{stat_id="60003",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| SSStore 每次写入平均耗时 | sum(rate(ob_sysstat{stat_id="60004",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="60003",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| SSStore 每秒写入数据量 | sum(rate(ob_sysstat{stat_id="60005",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | byte |
| 每秒写入平均耗时 | 1000000 * (avg(rate(node_disk_write_time_seconds_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) / (avg(rate(node_disk_writes_completed_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) | us |
| 平均每次 IO 写入耗时 | 1000000 * (avg(rate(node_disk_write_time_seconds_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) / (avg(rate(node_disk_writes_completed_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) | us |
| 过去1分钟系统平均负载 | avg(node_load1{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
  |  |
| 过去15分钟系统平均负载 | avg(node_load15{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
  |  |
| 过去5分钟系统平均负载 | avg(node_load5{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
  |  |
| 触发合并阈值 | sum(ob_sysstat{stat_id="130002",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| 内核 Buffer Cache 大小 | avg(node_memory_Buffers_bytes{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1073741824  | GB |
| 可用物理内存大小 | avg(node_memory_MemFree_bytes{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1073741824  | GB |
| 使用物理内存大小 | (avg(node_memory_MemTotal_bytes{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 - avg(node_memory_MemFree_bytes{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 - avg(node_memory_Cached_bytes{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 - avg(node_memory_Buffers_bytes{@LABELS}) by (@GBLABELS)) / 1073741824  | GB |
| MEMStore的limit | sum(ob_sysstat{stat_id="130004",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| MEMStore使用百分比 | 100 * sum(ob_sysstat{stat_id="130001",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / sum(ob_sysstat{stat_id="130004",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
  | % |
| 写锁等待失败次数 | sum(rate(ob_sysstat{stat_id="60022",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 写锁等待成功次数 | sum(rate(ob_sysstat{stat_id="60021",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 写锁平均等待耗时 | sum(rate(ob_sysstat{stat_id="60023",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / (sum(rate(ob_sysstat{stat_id="60021",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="60022",@LABELS}[@INTERVAL])) by (@GBLABELS))  | us |
| 每秒接收数据量 | avg(rate(node_network_receive_bytes_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| 每秒发送数据量 | avg(rate(node_network_transmit_bytes_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| CPU使用率 | 100 * sum(ob_sysstat{stat_id="140006",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / sum(ob_sysstat{stat_id="140005",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
  | % |
| 分区数量 | sum(ob_partition_num{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
  |  |
| 执行计划缓存命中率 | 100 * sum(rate(ob_plan_cache_hit_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_plan_cache_access_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | % |
| 执行计划缓存大小 | sum(ob_plan_cache_memory_bytes{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| 平均每秒 SQL 进等待队列的次数 | sum(rate(ob_sysstat{stat_id="20001",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| SQL 在等待队列中等待耗时 | sum(rate(ob_sysstat{stat_id="20002",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="20001",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 行缓存命中率 | 100 * 1 / (1 + sum(rate(ob_sysstat{stat_id="50001",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="50000",@LABELS}[@INTERVAL])) by (@GBLABELS))  | % |
| 缓存大小 | sum(ob_cache_size_bytes{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| Rpc 收包吞吐量 | sum(rate(ob_sysstat{stat_id="10001",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | byte |
| Rpc 收包平均耗时 | (sum(rate(ob_sysstat{stat_id="10005",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="10006",@LABELS}[@INTERVAL])) by (@GBLABELS)) / sum(rate(ob_sysstat{stat_id="10000",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| Rpc 发包吞吐量 | sum(rate(ob_sysstat{stat_id="10003",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | byte |
| Rpc 发包平均耗时 | (sum(rate(ob_sysstat{stat_id="10005",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="10006",@LABELS}[@INTERVAL])) by (@GBLABELS)) / sum(rate(ob_sysstat{stat_id="10002",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 每秒处理sql语句数 | sum(rate(ob_sysstat{stat_id="40002",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40004",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40006",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40008",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40000",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 服务端每条 SQL 语句平均处理耗时 | (sum(rate(ob_sysstat{stat_id="40003",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40005",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40007",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40009",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40001",@LABELS}[@INTERVAL])) by (@GBLABELS)) /(sum(rate(ob_sysstat{stat_id="40002",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40004",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40006",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40008",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 + sum(rate(ob_sysstat{stat_id="40000",@LABELS}[@INTERVAL])) by (@GBLABELS))  | us |
| 每秒处理 Delete 语句数 | sum(rate(ob_sysstat{stat_id="40008",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 服务端每条 Delete 语句平均处理耗时 | sum(rate(ob_sysstat{stat_id="40009",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="40008",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 每秒处理分布式执行计划数 | sum(rate(ob_sysstat{stat_id="40012",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 每秒处理 Insert 语句数 | sum(rate(ob_sysstat{stat_id="40002",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 服务端每条 Insert 语句平均处理耗时 | sum(rate(ob_sysstat{stat_id="40003",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="40002",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 每秒处理本地执行数 | sum(rate(ob_sysstat{stat_id="40010",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 每秒处理远程执行计划数 | sum(rate(ob_sysstat{stat_id="40011",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 每秒处理 Replace 语句数 | sum(rate(ob_sysstat{stat_id="40004",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 服务端每条 Replace 语句平均处理耗时 | sum(rate(ob_sysstat{stat_id="40005",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="40004",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 每秒处理 Select 语句数 | sum(rate(ob_sysstat{stat_id="40000",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 服务端每条 Select 语句平均处理耗时 | sum(rate(ob_sysstat{stat_id="40001",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="40000",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 每秒处理 Update 语句数 | sum(rate(ob_sysstat{stat_id="40006",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 服务端每条 Update 语句平均处理耗时 | sum(rate(ob_sysstat{stat_id="40007",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="40006",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 表数量 | max(ob_table_num{@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
  |  |
| MEMStore 总大小 | sum(ob_sysstat{stat_id="130001",@LABELS}) by ([@GBLABELS) ](/GBLABELS) )
 / 1048576  | MB |
| 每秒处理事务数 | sum(rate(ob_sysstat{stat_id="30005",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 服务端每个事务平均处理耗时 | sum(rate(ob_sysstat{stat_id="30006",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="30005",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 每秒提交的事务日志数 | sum(rate(ob_sysstat{stat_id="30002",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 每次事务日志网络同步平均耗时 | sum(rate(ob_sysstat{stat_id="30000",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_sysstat{stat_id="30001",@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | us |
| 每秒等待事件次数 | sum(rate(ob_waitevent_wait_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | times/s |
| 等待事件平均耗时 | sum(rate(ob_waitevent_wait_seconds_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
 / sum(rate(ob_waitevent_wait_total{@LABELS}[@INTERVAL])) by ([@GBLABELS) ](/GBLABELS) )
  | s |

