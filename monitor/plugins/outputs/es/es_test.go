package es

import (
	"bufio"
	"context"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	esutil "github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/monitor/message"
)

var content0 = `[2022-01-20 10:44:04.965609] INFO  [RPC.OBMYSQL] obsm_handler.cpp:239 [3452723][0][Y0-0000000000000000] [lt=28] [dc=0] unlock session of tenant(conn->version_=0, conn->sessid_=3221658203, proxy_sessid=0, conn->tenant_id_=1)
[2022-01-20 10:44:04.965620] INFO  [RPC.OBMYSQL] obsm_handler.cpp:253 [3452723][0][Y0-0000000000000000] [lt=7] [dc=0] connection close(easy_connection_str(c)="0.0.0.0_127.0.0.1:37641_-1_0x7fdf98367c20 tp=0 t=1642646644010220-1642646644464470 s=3 r=0 io=683/24747 sq=0", version=0, sessid=3221658203, proxy_sessid=0, tenant_id=1, server_id=1, from_proxy=false, from_java_client=false, c/s protocol="OB_MYSQL_CS_TYPE", is_need_clear_sessid_=true, ret=0)
[2022-01-20 10:44:04.965625] INFO  [SQL.SESSION] ob_sql_session_info.cpp:313 [3452276][1182][Y0-0000000000000000] [lt=14] [dc=0] end trans successfully(sessid=3221658203, proxy_sessid=0, version=0, trans id={hash:0, inc:0, addr:"0.0.0.0", t:0}, has_called_txs_end_trans=true)
[2022-01-20 10:44:04.965729] INFO  [SERVER] obmp_disconnect.cpp:79 [3452276][1182][Y0-0000000000000000] [lt=21] [dc=0] free session successfully(sessid=3221658203, version=0)
[2022-01-20 10:44:04.968678] INFO  [STORAGE.TRANS] ob_trans_ctx_mgr.cpp:1909 [3452591][1808][Y0-0000000000000000] [lt=28] [dc=0] mgr cache hit rate(partition={tid:1099511627925, partition_id:15, part_cnt:0}, hit_count=129902717575, access_count=153602511735, hit_rate=8.457069881660961297e-01)
[2022-01-20 10:44:04.970721] INFO  easy_connection.c:2268 [3452723][0][Y0-0000000000000000] [lt=17] [dc=0] Connection not in ESTABLISHED state, conn(127.0.0.1:2881_127.0.0.1:37639_2427_0x7fdf983842c0 tp=0 t=1642646644010016-1642646644470245 s=3 r=0 io=7374/140397 sq=0), status(3), ref(0), time(0.500454s).
[2022-01-20 10:44:04.970740] INFO  easy_connection.c:424 [3452723][0][Y0-0000000000000000] [lt=15] [dc=0] Destroying connection, conn(127.0.0.1:2881_127.0.0.1:37639_2427_0x7fdf983842c0 tp=0 t=1642646644010016-1642646644470245 s=3 r=0 io=7374/140397 sq=0), reason(on_timeout_conn).
[2022-01-20 10:44:04.970757] INFO  easy_connection.c:509 [3452723][0][Y0-0000000000000000] [lt=4] [dc=0] Socket closed, fd(2427), conn(127.0.0.1:2881_127.0.0.1:37639_2427_0x7fdf983842c0 tp=0 t=1642646644010016-1642646644470245 s=3 r=0 io=7374/140397 sq=0), ev_is_pending(0), ev_is_active(0), ev_timer_pending_addr(0x7fdf983784b0), ev_timer_pending(1), timeout_watcher(0x7fdf98384390).
[2022-01-20 10:44:04.970886] INFO  [RPC.OBMYSQL] obsm_handler.cpp:229 [3452723][0][Y0-0000000000000000] [lt=11] [dc=0] mark sessid unused(conn->version_=0, conn->sessid_=3221628863, proxy_sessid=0, server_id=1)
[2022-01-20 10:44:04.970907] INFO  [RPC.OBMYSQL] obsm_handler.cpp:239 [3452723][0][Y0-0000000000000000] [lt=20] [dc=0] unlock session of tenant(conn->version_=0, conn->sessid_=3221628863, proxy_sessid=0, conn->tenant_id_=1)
[2022-01-20 10:44:04.970917] INFO  [RPC.OBMYSQL] obsm_handler.cpp:253 [3452723][0][Y0-0000000000000000] [lt=5] [dc=0] connection close(easy_connection_str(c)="0.0.0.0_127.0.0.1:37639_-1_0x7fdf983842c0 tp=0 t=1642646644010016-1642646644470245 s=3 r=0 io=7374/140397 sq=0", version=0, sessid=3221628863, proxy_sessid=0, tenant_id=1, server_id=1, from_proxy=false, from_java_client=false, c/s protocol="OB_MYSQL_CS_TYPE", is_need_clear_sessid_=true, ret=0)
[2022-01-20 10:44:04.970918] INFO  [SQL.SESSION] ob_sql_session_info.cpp:313 [3452276][1182][Y0-0000000000000000] [lt=13] [dc=0] end trans successfully(sessid=3221628863, proxy_sessid=0, version=0, trans id={hash:0, inc:0, addr:"0.0.0.0", t:0}, has_called_txs_end_trans=true)
[2022-01-20 10:44:04.971011] INFO  [SERVER] obmp_disconnect.cpp:79 [3452276][1182][Y0-0000000000000000] [lt=22] [dc=0] free session successfully(sessid=3221628863, version=0)
[2022-01-20 10:44:04.973548] INFO  [STORAGE.TRANS] ob_xa_trans_relocate_worker.cpp:109 [3452613][1852][Y0-0000000000000000] [lt=37] [dc=0] XA transaction relocate task statistics(avg_time=5050)
[2022-01-20 10:44:04.983184] INFO  [CLOG] ob_log_state_driver_runnable.cpp:190 [3452682][1954][Y0-0000000000000000] [lt=26] [dc=0] ObLogStateDriverRunnable log delay histogram(histogram="Count: 0  Average: 0.0000  StdDev: 0.00
Min: 0.0000  Median: -nan  Max: 0.0000
------------------------------------------------------
")
[2022-01-20 10:44:04.985716] INFO  [SERVER] ob_inner_sql_connection.cpp:1329 [3452577][1784][Y0-0000000000000000] [lt=24] [dc=0] execute write sql(ret=0, tenant_id=1, affected_rows=1, sql="     update __all_weak_read_service set min_version=1642646644821354, max_version=1642646644821354     where level_id = 0 and level_value = '' and min_version = 1642646644770847 and max_version = 1642646644770847 ")
[2022-01-20 10:44:04.988553] INFO  [CLOG] ob_log_archive_and_restore_driver.cpp:249 [3452685][1960][YB42C0A8050C-0005D4D41E1DBA64] [lt=17] [dc=0] ObLogArchiveAndRestoreDriver round_cost_time(round_cost_time=5091, sleep_ts=494909, cycle_num=2528231)
[2022-01-20 10:49:14.186545] INFO  [STORAGE.TRANS] ob_gts_source.cpp:65 [3452590][1806][Y0-0000000000000000] [lt=8] [dc=0] gts statistics(tenant_id=1006, gts_rpc_cnt=0, get_gts_cache_cnt=0, get_gts_with_stc_cnt=0, try_get_gts_cache_cnt=0, try_get_gts_with_stc_cnt=0, wait_gts_elapse_cnt=0, try_wait_gts_elapse_cnt=0)
[2022-01-20 10:49:14.194210] INFO  [STORAGE] ob_partition_loop_worker.cpp:404 [3452695][1980][Y0-0000000000000000] [lt=17] [dc=0] gene checkpoint(pkey={tid:1099511627980, partition_id:8, part_cnt:0}, state=6, last_checkpoint=1642646954146683, cur_checkpoint=1642646954146683, last_max_trans_version=1641382401968431, max_trans_version=1641382401968431)
[2022-01-20 10:49:14.198285] INFO  [SERVER] ob_inner_sql_connection.cpp:1329 [3452577][1784][Y0-0000000000000000] [lt=17] [dc=0] execute write sql(ret=0, tenant_id=1, affected_rows=1, sql="     update __all_weak_read_service set min_version=1642646954046979, max_version=1642646954046979     where level_id = 0 and level_value = '' and min_version = 1642646953996617 and max_version = 1642646953996617 ")
[2022-01-20 10:49:14.224857] INFO  [SHARE] ob_server_blacklist.cpp:320 [3452071][782][Y0-0000000000000000] [lt=21] [dc=0] blacklist_loop exec finished(cost_time=48, is_enabled=true, dst_server count=0, send_cnt=0)
[2022-01-20 10:49:14.228370] INFO  [COMMON] ob_kvcache_store.cpp:811 [3451910][464][Y0-0000000000000000] [lt=15] [dc=0] Wash compute wash size(sys_total_wash_size=-2252341248, global_cache_size=2027823616, tenant_max_wash_size=1451155136, tenant_min_wash_size=0, tenant_ids_=[1, 500, 1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010], sys_cache_reserve_size=251658240, tg=time guard 'compute_tenant_wash_size' cost too much time, used=167, time_dist: 125,4,2,12,2,1)
[2022-01-20 10:49:14.228519] INFO  [COMMON] ob_kvcache_store.cpp:318 [3451910][464][Y0-0000000000000000] [lt=31] [dc=0] Wash time detail, (refresh_score_time=273, compute_wash_size_time=181, wash_sort_time=115, wash_time=2)
[2022-01-20 10:49:14.258981] INFO  [SERVER] ob_inner_sql_connection.cpp:1329 [3452577][1784][Y0-0000000000000000] [lt=22] [dc=0] execute write sql(ret=0, tenant_id=1, affected_rows=1, sql="     update __all_weak_read_service set min_version=1642646954096283, max_version=1642646954096283     where level_id = 0 and level_value = '' and min_version = 1642646954046979 and max_version = 1642646954046979 ")
[2022-01-20 10:49:14.260508] INFO  [CLOG.EXTLOG] ob_external_fetcher.cpp:276 [3452134][900][Y0-0000000000000000] [lt=23] [dc=0] [FETCH_LOG_STREAM] Wash Stream: wash expired stream success(count=0, retired_arr=[])
[2022-01-20 10:49:14.279415] INFO  [SERVER] ob_eliminate_task.cpp:199 [3452619][1864][Y0-0000000000000000] [lt=18] [dc=0] sql audit evict task end(evict_high_level=103079214, evict_batch_count=0, elapse_time=1, size_used=100, mem_used=16711680)
[2022-01-20 10:49:14.287539] INFO  [STORAGE.TRANS] ob_gts_source.cpp:65 [3452590][1806][Y0-0000000000000000] [lt=6] [dc=0] gts statistics(tenant_id=1007, gts_rpc_cnt=50, get_gts_cache_cnt=17353, get_gts_with_stc_cnt=82, try_get_gts_cache_cnt=0, try_get_gts_with_stc_cnt=0, wait_gts_elapse_cnt=82, try_wait_gts_elapse_cnt=0)
[2022-01-20 10:49:14.287562] INFO  [STORAGE.TRANS] ob_gts_source.cpp:65 [3452590][1806][Y0-0000000000000000] [lt=21] [dc=0] gts statistics(tenant_id=1007, gts_rpc_cnt=0, get_gts_cache_cnt=0, get_gts_with_stc_cnt=0, try_get_gts_cache_cnt=0, try_get_gts_with_stc_cnt=0, wait_gts_elapse_cnt=0, try_wait_gts_elapse_cnt=0)
[2022-01-20 10:49:14.287586] INFO  [STORAGE.TRANS] ob_gts_source.cpp:65 [3452590][1806][Y0-0000000000000000] [lt=7] [dc=0] gts statistics(tenant_id=1008, gts_rpc_cnt=50, get_gts_cache_cnt=17354, get_gts_with_stc_cnt=83, try_get_gts_cache_cnt=0, try_get_gts_with_stc_cnt=0, wait_gts_elapse_cnt=83, try_wait_gts_elapse_cnt=0)
[2022-01-20 10:49:14.287603] INFO  [STORAGE.TRANS] ob_gts_source.cpp:65 [3452590][1806][Y0-0000000000000000] [lt=16] [dc=0] gts statistics(tenant_id=1008, gts_rpc_cnt=0, get_gts_cache_cnt=0, get_gts_with_stc_cnt=0, try_get_gts_cache_cnt=0, try_get_gts_with_stc_cnt=0, wait_gts_elapse_cnt=0, try_wait_gts_elapse_cnt=0)
[2022-01-20 10:49:14.304463] INFO  [STORAGE] ob_partition_loop_worker.cpp:404 [3452695][1980][Y0-0000000000000000] [lt=30] [dc=0] gene checkpoint(pkey={tid:1099511627926, partition_id:6, part_cnt:0}, state=6, last_checkpoint=1642646954205497, cur_checkpoint=1642646954205497, last_max_trans_version=1641382402048600, max_trans_version=0)
[2022-01-20 10:49:14.312687] INFO  [STORAGE.TRANS] ob_tenant_weak_read_service.cpp:476 [3452617][1860][Y0-0000000000000000] [lt=23] [dc=0] [WRS] [TENANT_WEAK_READ_SERVICE] [STAT](tenant_id=1001, server_version={version:1642646954197066, total_part_count:476, valid_inner_part_count:121, valid_user_part_count:355}, server_version_delta=115608, in_cluster_service=true, cluster_version=1642646954096283, min_cluster_version=1642646954096283, max_cluster_version=1642646954096283, get_cluster_version_err=0, cluster_version_delta=216391, cluster_service_pkey={tid:1100611139404002, partition_id:0, part_cnt:0}, post_cluster_heartbeat_count=25135399, succ_cluster_heartbeat_count=53, cluster_heartbeat_interval=50000, local_cluster_version=0, local_cluster_delta=1642646954312674, force_self_check=false, weak_read_refresh_interval=50000)
[2022-01-20 10:49:14.319000] INFO  [SERVER] ob_inner_sql_connection.cpp:1329 [3452577][1784][Y0-0000000000000000] [lt=24] [dc=0] execute write sql(ret=0, tenant_id=1, affected_rows=1, sql="     update __all_weak_read_service set min_version=1642646954197066, max_version=1642646954197066     where level_id = 0 and level_value = '' and min_version = 1642646954096283 and max_version = 1642646954096283 ")
[2022-01-20 10:49:14.324126] INFO  [STORAGE] ob_freeze_info_snapshot_mgr.cpp:982 [3452169][970][Y0-0000000000000000] [lt=18] [dc=0] start reload freeze info and snapshots(is_remote_=true)
[2022-01-20 10:49:14.332262] INFO  [LIB] ob_json.cpp:278 [3451815][274][Y0-0000000000000000] [lt=14] [dc=0] invalid token type, maybe it is valid empty json type(cur_token_.type=93, ret=-5006)
[2022-01-20 10:49:14.332302] INFO  ob_config.cpp:956 [3451815][274][Y0-0000000000000000] [lt=21] [dc=0] succ to format_option_str(src="ASYNC NET_TIMEOUT = 30000000", dest="ASYNC NET_TIMEOUT  =  30000000")
[2022-01-20 10:49:14.332413] INFO  [SERVER] ob_remote_server_provider.cpp:208 [3451815][274][Y0-0000000000000000] [lt=8] [dc=0] [remote_server_provider] refresh server list(ret=0, ret="OB_SUCCESS", all_server_count=0)

`

var content1 = `[2022-01-20 10:44:55.740003] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=6] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:55.741387] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=16] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:55.741404] INFO  [RS] ob_rs_gts_monitor.cpp:232 [3452034][708][Y0-0000000000000000] [lt=16] [dc=0] check migrate unit finish
[2022-01-20 10:44:55.742923] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=6] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:55.742945] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=20] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:55.742952] INFO  [RS] ob_rs_gts_monitor.cpp:313 [3452034][708][Y0-0000000000000000] [lt=6] [dc=0] check shrink gts resource pool
[2022-01-20 10:44:56.039824] INFO  [SERVER] ob_inner_sql_connection.cpp:1329 [3452018][676][YB42C0A8050C-0005D4D502CECB7C] [lt=30] [dc=0] execute write sql(ret=0, tenant_id=1, affected_rows=2, sql="INSERT INTO __all_core_table (table_name, row_id, column_name, column_value) VALUES  ('__all_global_stat', 1, 'snapshot_gc_ts', '1642646691037115') ON DUPLICATE KEY UPDATE column_value = values(column_value)")
[2022-01-20 10:44:56.039903] INFO  [RS] ob_freeze_info_manager.cpp:1605 [3452018][676][YB42C0A8050C-0005D4D502CECB7C] [lt=24] [dc=0] get max frozen status for daily merge(ret=0, frozen_status={frozen_version:129, frozen_timestamp:1642615210776508, cluster_version:8590065741}, cost=2)
[2022-01-20 10:44:56.059613] INFO  [RS] ob_backup_lease_service.cpp:391 [3452043][726][Y0-0000000000000000] [lt=22] [dc=0] set_backup_leader_info_(ret=0, lease_info_={is_leader:true, lease_start_ts:1642646696059609, leader_epoch:1641382400000000, leader_takeover_ts:1641382402997190, round:1})
[2022-01-20 10:44:56.339561] INFO  [RS] ob_server_table_operator.cpp:379 [3451678][4][Y0-0000000000000000] [lt=15] [dc=0] svr_status(svr_status="active", display_status=1)
[2022-01-20 10:44:56.744800] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=6] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:56.744827] INFO  [RS] ob_rs_gts_monitor.cpp:174 [3452034][708][Y0-0000000000000000] [lt=24] [dc=0] distribute gts unit for server status change
[2022-01-20 10:44:56.746206] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=6] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:56.747538] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=17] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:56.747555] INFO  [RS] ob_rs_gts_monitor.cpp:232 [3452034][708][Y0-0000000000000000] [lt=15] [dc=0] check migrate unit finish
[2022-01-20 10:44:56.748883] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=5] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:56.748900] WARN  [RS] get_pool_ids_of_tenant (ob_unit_manager.cpp:2353) [3452034][708][Y0-0000000000000000] [lt=15] [dc=0] tenant doesn't own any pool(tenant_id=2, ret=0)
[2022-01-20 10:44:56.748905] INFO  [RS] ob_rs_gts_monitor.cpp:313 [3452034][708][Y0-0000000000000000] [lt=4] [dc=0] check shrink gts resource pool
[2022-01-20 10:44:57.040695] INFO  [RS] ob_freeze_info_manager.cpp:1605 [3452018][676][YB42C0A8050C-0005D4D502CECB7D] [lt=16] [dc=0] get max frozen status for daily merge(ret=0, frozen_status={frozen_version:129, frozen_timestamp:1642615210776508, cluster_version:8590065741}, cost=2)
[2022-01-20 10:44:57.062988] INFO  [RS] ob_backup_lease_service.cpp:391 [3452043][726][Y0-0000000000000000] [lt=26] [dc=0] set_backup_leader_info_(ret=0, lease_info_={is_leader:true, lease_start_ts:1642646697062984, leader_epoch:1641382400000000, leader_takeover_ts:1641382402997190, round:1})
`

var content2 = `[2022-01-20 10:37:56.377579] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=27] ObLogItemFactory statistics(alloc_count=74984, release_count=74983, used=0)
[2022-01-20 10:37:56.377606] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=35] ObElectionGCThread loop cost(cost time=0)
[2022-01-20 10:38:56.378092] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=35] ObLogItemFactory statistics(alloc_count=74986, release_count=74985, used=0)
[2022-01-20 10:38:56.378111] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=41] ObElectionGCThread loop cost(cost time=0)
[2022-01-20 10:39:56.378555] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=24] ObLogItemFactory statistics(alloc_count=74988, release_count=74987, used=0)
[2022-01-20 10:39:56.378578] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=32] ObElectionGCThread loop cost(cost time=1057)
[2022-01-20 10:40:56.379033] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=29] ObLogItemFactory statistics(alloc_count=74990, release_count=74989, used=0)
[2022-01-20 10:40:56.379056] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=38] ObElectionGCThread loop cost(cost time=0)
[2022-01-20 10:41:56.379506] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=32] ObLogItemFactory statistics(alloc_count=74992, release_count=74991, used=0)
[2022-01-20 10:41:56.379531] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=41] ObElectionGCThread loop cost(cost time=1055)
[2022-01-20 10:42:56.379982] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=33] ObLogItemFactory statistics(alloc_count=74994, release_count=74993, used=0)
[2022-01-20 10:42:56.380005] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=41] ObElectionGCThread loop cost(cost time=1056)
[2022-01-20 10:43:56.380445] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=28] ObLogItemFactory statistics(alloc_count=74996, release_count=74995, used=0)
[2022-01-20 10:43:56.380467] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=37] ObElectionGCThread loop cost(cost time=0)
[2022-01-20 10:44:56.380948] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=28] ObLogItemFactory statistics(alloc_count=74998, release_count=74997, used=0)
[2022-01-20 10:44:56.380971] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=36] ObElectionGCThread loop cost(cost time=1060)
[2022-01-20 10:45:56.381420] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=28] ObLogItemFactory statistics(alloc_count=75000, release_count=74999, used=0)
[2022-01-20 10:45:56.381438] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=42] ObElectionGCThread loop cost(cost time=1054)
[2022-01-20 10:46:56.381895] INFO  [ELECT] alloc (ob_election_async_log.cpp:72) [3452607][Y0-0000000000000000] [log=25] ObLogItemFactory statistics(alloc_count=75002, release_count=75001, used=0)
[2022-01-20 10:46:56.381914] INFO  [ELECT] run1 (ob_election_gc_thread.cpp:98) [3452607][Y0-0000000000000000] [log=32] ObElectionGCThread loop cost(cost time=0)
`

var pattern = regexp.MustCompile(`^\[(?P<time>[\d:. -]+)\]\s+(?P<level>[A-Z]+)\s+(?:\[(?P<module>[A-Z._]+)\]\s+)?(?:(?P<func>\w+) \((?P<file1>[\w]+\.\w+):(?P<line1>\d+)\)\s+|(?P<file2>[\w]+\.\w+):(?P<line2>\d+)\s+)?(?:\[(?P<thread>\d+)\](?:\[(\d+)\])?\[(?P<trace>[\w-]+)\]\s+)(?:\[[\w=]+\]\s+)*(?P<content>.+)`)

func TestWrite(t *testing.T) {
	//t.Skip()
	ctx := context.Background()
	out, err := NewESOutput(map[string]interface{}{
		"clientAddresses":  "127.0.0.1:9200",
		"indexNamePattern": "ocp_log_%Y%m%d",
		"batchSizeInBytes": 1048576,
		"maxBatchWait":     "1s",
		"docMap": map[string]interface{}{
			"timestamp":          "timestamp",
			"timestampPrecision": "1us",
			"name":               "file",
			"tags": map[string]string{
				"level": "level",
				"app":   "app",
			},
			"fields": map[string]string{
				"tags":    "tags",
				"content": "content",
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}
	ch := make(chan *message.Message, 100)
	go func() {
		//f, _ := os.Open("/tmp/log/observer.log.20220127071625")
		f, _ := os.Open("/tmp/log/rootservice.log.20220126173331")

		parseObLog("rootservice.log", f, ch)
		//parseObLog(strings.NewReader(content0), ch)
		//parseObLog(strings.NewReader(content1), ch)
		//parseObLog(strings.NewReader(content2), ch)
		close(ch)
	}()

	for msg := range ch {
		msg.AddTag("app", "ob")
		_ = out.Write(ctx, []*message.Message{msg})
		//fmt.Println(msg)
	}
	time.Sleep(5 * time.Second)
}

func parseObLog(name string, reader io.Reader, out chan<- *message.Message) error {
	scanner := bufio.NewScanner(reader)
	var msg *message.Message = nil
	content := make([]string, 0, 8)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		match := pattern.FindStringSubmatch(line)
		if match == nil {
			if msg != nil {
				content = append(content, line)
			}
			continue
		}
		if msg != nil {
			msg.AddField("content", strings.Join(content, "\n"))
			out <- msg
		}

		t, err := time.ParseInLocation("2006-01-02 15:04:05.999999999", match[pattern.SubexpIndex("time")], time.Local)
		if err != nil {
			continue
		}
		tags := make([]string, 0, 4)
		msg = message.NewMessage(name, message.Log, t).
			AddTag("level", strings.ToLower(match[pattern.SubexpIndex("level")]))
		content = content[0:0]
		content = append(content, match[pattern.SubexpIndex("content")])
		for _, name := range pattern.SubexpNames() {
			i := pattern.SubexpIndex(name)
			if i < 0 || i >= len(match) {
				continue
			}
			value := match[i]
			if value == "" {
				continue
			}
			switch name {
			case "level", "time", "content", "line1", "line2":
				continue
			default:
				if name == "file1" || name == "file2" {
					name = "file"
				}
				tags = append(tags, name+":"+value)
			}
		}
		msg.AddField("tags", tags)
	}
	if msg != nil {
		out <- msg
	}
	return scanner.Err()
}

func TestESOutput_toDocMap(t *testing.T) {
	config := DocMap{
		Timestamp:          "time",
		Name:               "name",
		TimestampPrecision: time.Second,
		Tags: map[string]string{
			"tag1": "t1",
			"tag2": "t2",
		},
		Fields: map[string]string{
			"field1": "f1",
			"field2": "f2",
		},
	}
	tmpDoc := make(map[string]interface{})

	t1 := time.Date(2022, 1, 2, 11, 12, 13, 0, time.Local)
	msg := message.NewMessage("test1", message.Log, t1).
		AddTag("tag1", "t1").
		AddField("field1", "f1")

	toDocMap(config, msg, tmpDoc)
	if tmpDoc["time"] != t1.Unix() {
		t.Error("time wrong")
	}
	if tmpDoc["name"] != "test1" {
		t.Error("name wrong")
	}
	if tmpDoc["t1"] != "t1" {
		t.Error("tag wrong")
	}
	if tmpDoc["f1"] != "f1" {
		t.Error("field wrong")
	}

	msg = message.NewMessage("test2", message.Log, t1).
		AddTag("tag2", "t2").
		AddField("field2", "f2")
	toDocMap(config, msg, tmpDoc)

	if tmpDoc["time"] != t1.Unix() {
		t.Error("time wrong")
	}
	if tmpDoc["name"] != "test2" {
		t.Error("name wrong")
	}
	if tmpDoc["t1"] != nil {
		t.Error("tag wrong")
	}
	if tmpDoc["f1"] != nil {
		t.Error("field wrong")
	}
	if tmpDoc["t2"] != "t2" {
		t.Error("tag wrong")
	}
	if tmpDoc["f2"] != "f2" {
		t.Error("field wrong")
	}
}

type MockIndexer struct {
	items []esutil.BulkIndexerItem
}

func (m *MockIndexer) Add(ctx context.Context, item esutil.BulkIndexerItem) error {
	m.items = append(m.items, item)
	return nil
}

func (m *MockIndexer) Close(context.Context) error {
	return nil
}

func (m *MockIndexer) Stats() esutil.BulkIndexerStats {
	return esutil.BulkIndexerStats{}
}

func TestESOutput_Write(t *testing.T) {
	es := &ESOutput{
		config: Config{
			IndexNamePattern: "idx_%Y%m%d%H",
			DocMap: DocMap{
				Timestamp:          "time",
				Name:               "name",
				TimestampPrecision: time.Microsecond,
				Tags: map[string]string{
					"tag1": "t1",
					"tag2": "t2",
				},
				Fields: map[string]string{
					"field1": "f1",
					"field2": "f2",
				},
			},
		},
		indexer: &MockIndexer{},
	}

	t1 := time.Date(2022, 1, 2, 11, 12, 13, 122223000, time.Local)
	msg := message.NewMessage("test1", message.Log, t1).
		AddTag("tag1", "t1").
		AddField("field1", "f1")

	err := es.Write(context.Background(), []*message.Message{
		msg,
	})
	if err != nil {
		t.Error("write fail", err)
	}
	if len(es.indexer.(*MockIndexer).items) != 1 {
		t.Error("write wrong")
	}
	es.Description()
	es.SampleConfig()
}

func TestESOutput_indexName(t *testing.T) {
	t1 := time.Date(2022, 1, 2, 11, 12, 13, 0, time.Local)
	es := &ESOutput{
		config: Config{
			IndexNamePattern: "idx_%Y%m%d%H",
		},
	}
	name := es.indexName(t1)
	if name != "idx_2022010211" {
		t.Error("index name wrong")
	}
}

func Test_toESConfig(t *testing.T) {
	esConfig := toESConfig(Config{
		ClientAddresses: "127.0.0.1:9200,127.0.0.1:9200",
	})
	if esConfig.Addresses[0] != "http://127.0.0.1:9200" || esConfig.Addresses[1] != "http://127.0.0.1:9200" {
		t.Error("es config wrong")
	}
}

func Test_toBulkItem(t *testing.T) {
	es := &ESOutput{
		stat: Stat{},
	}
	item := es.toBulkItem("idx1", nil, []byte("{}"))
	if item.Action != "index" || item.Index != "idx1" {
		t.Error("bulkitem wrong")
	}
}

func TestESOutput_getRouting(t *testing.T) {
	es := &ESOutput{
		config: Config{
			RoutingField: "traceId",
		},
	}
	msg := &message.Message{}
	msg.AddTag("traceId", "YYYYYAAAA1111")
	routing := es.getRouting(msg)
	assert.Equal(t, *routing, "YYYYYAAAA1111")
}
