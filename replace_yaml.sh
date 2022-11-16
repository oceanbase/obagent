#!/bin/bash
### 关闭这个功能
http_basic_auth_user=${HTTP_BASIC_AUTH_USER:-'admin'}
http_basic_auth_password=${HTTP_BASIC_AUTH_PASSWORD:-'root'}
pprof_basic_auth_user=${PPROF_BASIC_AUTH_USER:-'admin'}
pprof_basic_auth_password=${PPROF_BASIC_AUTH_PASSWORD:-'root'}


monitor_user=$MONITOR_USER
monitor_password=$MONITOR_PASSWORD

sql_port=${SQL_PORT:-'2881'}
rpc_port=${RPC_PORT:-'2882'}

ob_install_path=${OB_INSTALL_PATH:-'ori_path'}
host_ip=$HOST_IP

cluster_name=$CLUSTER_NAME
cluster_id=$CLUSTER_ID

zone_name=$ZONE_NAME
ob_monitor_status=${OB_MONITOR_STATUS:-'active'}
ob_log_monitor_status=${OB_LOG_MONITOR_STATUS:-'inactive'}
host_monitor_status=${HOST_MONITOR_STATUS:-'active'}
alertmanager_address=${ALERTMANAGER_ADDRESS:-'temp'}
disable_http_basic_auth=${DISABLE_HTTP_BASIC_AUTH:-'true'}
disable_pprof_basic_auth=${DISABLE_PPROF_BASIC_AUTH:-'true'}

### 配置 monagent_basic_auth.yaml
sed -i "s/{http_basic_auth_user}/${http_basic_auth_user}/g" ./conf/config_properties/monagent_basic_auth.yaml
sed -i "s/{http_basic_auth_password}/${http_basic_auth_password}/g" ./conf/config_properties/monagent_basic_auth.yaml
sed -i "s/{pprof_basic_auth_user}/${pprof_basic_auth_user}/g" ./conf/config_properties/monagent_basic_auth.yaml
sed -i "s/{pprof_basic_auth_password}/${pprof_basic_auth_password}/g" ./conf/config_properties/monagent_basic_auth.yaml

### 配置 monagent_pipeline.yaml
sed -i "s/{monitor_user}/${monitor_user}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{monitor_password}/${monitor_password}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{sql_port}/${sql_port}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{rpc_port}/${rpc_port}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{ob_install_path}/${ob_install_path}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{host_ip}/${host_ip}/g" ./conf/config_properties/monagent_pipeline.yaml

sed -i "s/{cluster_name}/${cluster_name}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{cluster_id}/${cluster_id}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{zone_name}/${zone_name}/g" ./conf/config_properties/monagent_pipeline.yaml

sed -i "s/{ob_monitor_status}/${ob_monitor_status}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{ob_log_monitor_status}/${ob_log_monitor_status}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{host_monitor_status}/${host_monitor_status}/g" ./conf/config_properties/monagent_pipeline.yaml
sed -i "s/{alertmanager_address}/${alertmanager_address}/g" ./conf/config_properties/monagent_pipeline.yaml

###  替换 ip
sed -i 's/127.0.0.1/${monagent.host.ip}/g' ./conf/module_config/monitor_ob.yaml
sed -i "s/{disable_http_basic_auth}/${disable_http_basic_auth}/g" ./conf/module_config/monagent_basic_auth.yaml
sed -i "s/{disable_pprof_basic_auth}/${disable_pprof_basic_auth}/g" ./conf/module_config/monagent_basic_auth.yaml