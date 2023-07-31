# 输出插件（Exporter）指标

输出插件（Exporter）暴露的指标见下表：

## 主机指标

| **指标名** | **Label**                      | **描述** | **类型** |
| --- |--------------------------------| --- | --- |
| node_cpu_seconds_total | cpu,mode,svr_ip                | CPU 时间 | counter |
| node_disk_read_bytes_total | device,svr_ip                  | 磁盘读取字节数 | counter |
| node_disk_read_time_seconds_total | device,svr_ip                  | 磁盘读取消耗总时间 | counter |
| node_disk_reads_completed_total | device,svr_ip                  | 磁盘读取完成次数 | counter |
| node_disk_written_bytes_total | device,svr_ip                  | 磁盘写入字节数 | counter |
| node_disk_write_time_seconds_total | device,svr_ip                  | 磁盘写入消耗总时间 | counter |
| node_disk_writes_completed_total | device,svr_ip                  | 磁盘写入完成次数 | counter |
| node_disk_io_time_weighted_seconds_total | device,svr_ip                  | 磁盘IO时间 | counter |
| node_filesystem_avail_bytes | device,fstype,mountpoint,svr_ip | 文件系统可用大小 | gauge |
| node_filesystem_size_bytes | device,fstype,mountpoint,svr_ip | 文件系统大小 | gauge |
| node_load1 | svr_ip                         | 1 分钟平均 load | gauge |
| node_load5 | svr_ip                         | 5 分钟平均 load | gauge |
| node_load15 | svr_ip                         | 15 分钟平均 load | gauge |
| node_memory_Buffers_bytes | svr_ip                         | 内存 buffer 大小 | gauge |
| node_memory_Cached_bytes | svr_ip                         | 内存 cache 大小 | gauge |
| node_memory_MemFree_bytes | svr_ip                         | 内存 free 大小 | gauge |
| node_memory_SReclaimable_bytes | svr_ip                         | 可回收slab内存的大小 | gauge |
| node_memory_MemTotal_bytes | svr_ip                         | 内存总大小 | gauge |
| node_network_receive_bytes_total | device,svr_ip                  | 网络接受总字节数 | counter |
| node_network_transmit_bytes_total | device,svr_ip                  | 网络发送总字节数 | counter |
| node_ntp_offset_seconds | svr_ip                         | NTP 时钟偏移 | gauge |
| cpu_count | svr_ip                         | cpu核数 | gauge |
| node_net_bandwidth_bps | device,svr_ip                  | 网卡速率 | gauge |
| io_util | device,svr_ip | IO负载 | gauge |
| io_await | device,svr_ip | IO耗时 | gauge |


## OceanBase 数据库指标

| **指标名**                                    | **label**                                                                       | **含义**             | **类型** |
|--------------------------------------------|---------------------------------------------------------------------------------|--------------------| --- |
| ob_active_session_num                      | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name            | 活跃连接数              | gauge |
| ob_all_session_num                         | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name            | 总连接数               | gauge |
| ob_cache_size_bytes                        | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name,cache_name | kvcache 大小         | gauge |
| ob_server_num                              | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | observer数量         | gauge |
| ob_partition_num                           | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name            | 分区数                | gauge |
| ob_plan_cache_access_total                 | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | 执行计划访问次数           | counter |
| ob_plan_cache_hit_total                    | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | 执行计划命中次数           | counter |
| ob_plan_cache_memory_bytes                 | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | plancache大小        | gauge |
| ob_table_num                               | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | 表数量                | gauge |
| ob_waitevent_wait_seconds_total            | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | 等待事件总等待时间          | counter |
| ob_waitevent_wait_total                    | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | 等待事件总等待次数          | counter |
| ob_system_event_total_waits                | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name,event_group             | 系统事件总等待次数          | counter |
| ob_system_event_time_waited                | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name,event_group             | 系统事件总等待时间          | counter |
| ob_disk_free_bytes                         | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | OceanBase 磁盘剩余大小   | gauge |
| ob_disk_total_bytes                        | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | OceanBase 磁盘总大小    | gauge |
| ob_memstore_active_bytes                   | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | 活跃 memstore 大小     | gauge |
| ob_memstore_freeze_times                   | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | memstore 冻结次数      | counter |
| ob_memstore_freeze_trigger_bytes           | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | memstore 冻结阈值      | gauge |
| ob_memstore_total_bytes                    | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | memstore 总大小       | gauge |
| ob_server_resource_cpu                     | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | observer可用cpu数     | gauge |
| ob_server_resource_cpu_assigned            | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | observer 已分配 CPU 数 | gauge |
| ob_server_resource_memory_bytes            | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | observer 可用内存大小    | gauge |
| ob_server_resource_memory_assigned_bytes   | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | observer 已分配内存大小   | gauge |
| ob_server_resource_disk_bytes              | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | observer可用磁盘大小     | gauge |
| ob_server_resource_cpu_assigned_percent    | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | observer CPU使用率    | gauge |
| ob_server_resource_memory_assigned_percent | ob_cluster_id,ob_cluster_name,obzone,svr_ip                                     | observer 内存使用率     | gauge |
| ob_tenant_resource_max_cpu                  | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                        | 租户最大可用cpu数     | gauge |
| ob_tenant_resource_min_cpu          | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | 租户最小可用cpu数 | gauge |
| ob_tenant_resource_max_memory                  | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                        | 租户最大可用内存大小     | gauge |
| ob_tenant_resource_min_memory          | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | 租户最小可用内存大小 | gauge |
| ob_tenant_assigned_cpu_total                  | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                        | observer cpu总数     | gauge |
| ob_tenant_assigned_cpu_assigned         | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | observer已分配cpu数 | gauge |
| ob_tenant_assigned_mem_total                  | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                        | observer内存总量     | gauge |
| ob_tenant_assigned_mem_assigned          | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | observer已分配内存量 | gauge |
| ob_tenant_disk_data_size          | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | observer数据盘使用量 | gauge |
| ob_tenant_disk_log_size          | ob_cluster_id,ob_cluster_name,obzone,svr_ip,ob_tenant_id,tenant_name                         | observer日志盘使用量 | gauge |
| ob_disk_total_bytes          | ob_cluster_id,ob_cluster_name,obzone,svr_ip                         | observer磁盘总量 | gauge |
| ob_disk_free_bytes          | ob_cluster_id,ob_cluster_name,obzone,svr_ip                         | observer空闲磁盘量 | gauge |
| ob_unit_num                                | ob_cluster_id,ob_cluster_name,obzone,svr_ip                         | observer unit 数量   | gauge |
| ob_sysstat                                 | ob_cluster_id,ob_cluster_name,obzone,svr_ip,tenant_name,stat_id     | ob内部统计项            | 不同 stat_id 不相同，参考对应部分解释 |
