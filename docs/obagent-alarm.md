# OBAgent 报警项

| 报警项 | 监控指标 | 阈值
| --- | --- | --- |
| ob_host_connection_percent_over_threshold | 100 * max(ob_active_session_num{@LABELS} / 262144) by (@GBLABELS) | 80 |
| ob_host_cpu_percent |  100 * (1 - sum(rate(node_cpu_seconds_total{mode="idle", @LABELS}[@INTERVAL])) by (@GBLABELS) / sum(rate(node_cpu_seconds_total{@LABELS}[@INTERVAL])) by (@GBLABELS)) | 100  |
| ob_cpu_percent_over_threshold | 100 * sum(ob_sysstat{stat_id="140006",@LABELS}) by (@GBLABELS) / sum(ob_sysstat{stat_id="140005",@LABELS}) by (@GBLABELS) | 90  |
| ob_host_disk_percent_over_threshold | 100 * (1 - avg(node_filesystem_avail_bytes{@LABELS}) by (@GBLABELS) / avg(node_filesystem_size_bytes{@LABELS}) by (@GBLABELS)) | 97  |
| ob_cluster_frozen_version_delta_over_threshold | max(ob_zone_stat{name="frozen_version",@LABELS}) by (@GBLABELS) - min(ob_zone_stat{name="last_merged_version",@LABELS}) by (@GBLABELS) | 1  |
| ob_host_net_recv_percent_over_threshold | 100 * max(sum(rate(node_network_receive_bytes_total{@LABELS}[@INTERVAL])) by (device,@GBLABELS) / sum(bandwidth{@LABELS})) by (@GBLABELS) | 80  |
| ob_host_net_send_percent_over_threshold | 100 * max(sum(rate(node_network_transmit_bytes_total{@LABELS}[@INTERVAL])) by (device,@GBLABELS) / sum(bandwidth{@LABELS})) by (@GBLABELS) | 80  |
| ob_cluster_exists_inactive_server | max(ob_server_num{status="inactive",@LABELS}) by (@GBLABELS) | 0  |
| ob_cluster_exists_index_fail_table | sum(ob_index_error_num{@LABELS}) by (@GBLABELS) | 0  |
| ob_host_load1_per_cpu_over_threshold | sum(node_load1{@LABELS}) by (@GBLABELS) / sum(cpu_count{@LABELS}) by (@GBLABELS) | 2  |
| ob_host_mem_percent_over_threshold | (1 - (avg(node_memory_MemFree_bytes{@LABELS}) by (@GBLABELS) + avg(node_memory_Cached_bytes{@LABELS}) by (@GBLABELS) + avg(node_memory_Buffers_bytes{@LABELS}) by (@GBLABELS)) / avg(node_memory_MemTotal_bytes{@LABELS}) by (@GBLABELS)) * 100 | 90  |
| ob_cluster_merge_timeout | max(ob_zone_stat{name="is_merge_timeout",@LABELS}) by (@GBLABELS) | 1  |
| ob_cluster_merge_error | max(ob_zone_stat{name="is_merge_error",@LABELS}) by (@GBLABELS) | 1  |
| ob_host_partition_count_over_threshold |  sum(ob_partition_num{@LABELS}) by (@GBLABELS) | 30000  |
| ob_host_disk_readonly | max(node_filesystem_readonly{@LABELS}) by (@GBLABELS) | 1  |
| ob_server_sstable_percent_over_threshold | 100 * (sum(ob_disk_total_bytes{@LABELS}) by (@GBLABELS) - sum(ob_disk_free_bytes{@LABELS}) by (@GBLABELS)) / sum(ob_disk_total_bytes{@LABELS}) by (@GBLABELS) | 85  |
| tenant_active_memstore_percent_over_threshold | 100 * sum(ob_sysstat{stat_id="130000",@LABELS}) by (@GBLABELS) / sum(ob_sysstat{stat_id="130002",@LABELS}) by (@GBLABELS) | 110  |
| tenant_memstore_percent_over_threshold | 100 * sum(ob_sysstat{stat_id="130001",@LABELS}) by (@GBLABELS) / sum(ob_sysstat{stat_id="130004",@LABELS}) by (@GBLABELS) | 85  |
