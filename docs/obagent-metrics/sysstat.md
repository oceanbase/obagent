# 系统统计项

系统统计项（sysstat）的信息见下表：

| **stat_id** | **描述** | **类型** |
| --- | --- | --- |
| 10000 | 收到的 RPC 包数量 | counter |
| 10001 | 收到的 RPC 包大小 | counter |
| 10002 | 发送的 RPC 包数量 | counter |
| 10003 | 发送的 RPC 包大小 | counter |
| 10005 |  RPC 网络延迟 | counter |
| 10006 |  RPC 框架延迟 | counter |
| 20001 | 请求出队列次数 | counter |
| 20002 | 请求在队列中时间 | counter |
| 30000 | clog 同步时间 | counter |
| 30001 | clog 同步次数 | counter |
| 30002 | clog 提交次数 | counter |
| 30005 | 事务数 | counter |
| 30006 | 事务总时间 | counter |
| 40000 | select SQL 数 | counter |
| 40001 | select SQL执行时间 | counter |
| 40002 | insert SQL 数 | counter |
| 40003 | insert SQL 执行时间 | counter |
| 40004 | replace SQL 数 | counter |
| 40005 | replace SQL 执行时间 | counter |
| 40006 | update SQL 数 | counter |
| 40007 | update SQL 执行时间 | counter |
| 40008 | delete SQL 数 | counter |
| 40009 | delete SQL 执行时间 | counter |
| 40010 | 本地 SQL 执行次数 | counter |
| 40011 | 远程 SQL 执行次数 | counter |
| 40012 | 分布式 SQL 执行次数 | counter |
| 50000 | row cache 命中次数 | counter |
| 50001 | row cache 没有命中次数 | counter |
| 50008 | block cache 命中次数 | counter |
| 50009 | block cache 没有命中次数 | counter |
| 60000 | IO 读次数 | counter |
| 60001 | IO 读延时 | counter |
| 60002 | IO 读字节数 | counter |
| 60003 | IO 写次数 | counter |
| 60004 | IO 写延时 | counter |
| 60005 | IO 写字节数 | counter |
| 60019 | memstore 读锁成功次数 | counter |
| 60020 | memstore 读锁失败次数 | counter |
| 60021 | memstore 写锁成功次数 | counter |
| 60022 | memstore 写锁成功次数 | counter |
| 60023 | memstore 等写锁时间 | counter |
| 60024 | memstore 等读锁时间 | counter |
| 80040 | clog写次数 | counter |
| 80041 | clog写时间 | counter |
| 80057 | clog大小 | counter |
| 130000 | 活跃 memstore 大小 | gauge |
| 130001 | memstore 总大小 | gauge |
| 130002 | 触发 major freeze 阈值 | gauge |
| 130004 | memstore 大小限制 | gauge |
| 140002 | 最大可使用内存 | gauge |
| 140003 | 已使用内存 | gauge |
| 140005 | 最大可使用 CPU | gauge |
| 140006 | 已使用  CPU | gauge |
