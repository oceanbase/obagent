package common

// for ob 4.0
const obVersion4 = "4.0.0.0"

const (
	showObVersion           = "show variables like 'version_comment'"
	selectHostIp            = "select host_ip()"
	selectRpcPort           = "select rpc_port()"
	selectObserverId        = "select id from __all_server where svr_ip = ? and svr_port = ?"
	selectAllTenants        = "select tenant_id from gv$unit where ((svr_ip = ? and svr_port = ?) or (migrate_from_svr_ip = ? and migrate_from_svr_port = ?)) union select 1 order by tenant_id"
	selectObserverStartTime = "select start_service_time from __all_server where svr_ip = ? and svr_port = ?"

	//for ob 4.0
	selectObserverIdForObVersion4        = "select id from DBA_OB_SERVERS where svr_ip = ? and svr_port = ?"
	selectObserverStartTimeForObVersion4 = "select floor(unix_timestamp(start_service_time) *1000000) as start_service_time from DBA_OB_SERVERS where svr_ip = ? and svr_port = ?"
	selectAllTenantsForObVersion4        = "select tenant_id from GV$OB_UNITS where (svr_ip = ? and svr_port = ?) union select 1 order by tenant_id"
)
