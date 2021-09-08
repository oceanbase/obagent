// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package prometheus

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/oceanbase/obagent/metric"
)

func TestPrometheus_Collect(t *testing.T) {
	config := `{"urls":["http://127.0.0.1:9100/metrics"]}`

	go func() {
		http.HandleFunc("/metrics", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte(tempJson))
		})
		http.ListenAndServe(":9100", nil)
	}()
	time.After(time.Millisecond * 100)
	sourceConfig := make(map[string]interface{})
	err := json.Unmarshal([]byte(config), &sourceConfig)
	if err != nil {
		t.Errorf("json Unmarshal err %s", err.Error())
	}
	p := &Prometheus{}
	err = p.Init(sourceConfig)
	if err != nil {
		t.Errorf("input prometheus init err %s", err.Error())
	}

	tests := []struct {
		name    string
		fields  Prometheus
		want    []metric.Metric
		wantErr bool
	}{
		{name: "test1", fields: *p, want: nil, wantErr: false},
		{name: "test2", fields: *p, want: nil, wantErr: false},
		{name: "test3", fields: *p, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Prometheus{
				sourceConfig: tt.fields.sourceConfig,
				config:       tt.fields.config,
				httpClient:   tt.fields.httpClient,
			}
			_, err := p.Collect()
			if (err != nil) != tt.wantErr {
				t.Errorf("Collect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

var tempJson = `# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 0
go_gc_duration_seconds{quantile="1"} 0
go_gc_duration_seconds_sum 0
go_gc_duration_seconds_count 0
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 7
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.16.5"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 1.826528e+06
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 1.826528e+06
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 1.444728e+06
# HELP go_memstats_frees_total Total number of frees.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 894
# HELP go_memstats_gc_cpu_fraction The fraction of this program's available CPU time used by the GC since the program started.
# TYPE go_memstats_gc_cpu_fraction gauge
go_memstats_gc_cpu_fraction 0
# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 4.116344e+06
# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 1.826528e+06
# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used.
# TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 6.3135744e+07
# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.
# TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 3.31776e+06
# HELP go_memstats_heap_objects Number of allocated objects.
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 15983
# HELP go_memstats_heap_released_bytes Number of heap bytes released to OS.
# TYPE go_memstats_heap_released_bytes gauge
go_memstats_heap_released_bytes 6.303744e+07
# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 6.6453504e+07
# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 0
# HELP go_memstats_lookups_total Total number of pointer lookups.
# TYPE go_memstats_lookups_total counter
go_memstats_lookups_total 0
# HELP go_memstats_mallocs_total Total number of mallocs.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 16877
# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures.
# TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 14400
# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system.
# TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 16384
# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures.
# TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 81056
# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system.
# TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 81920
# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place.
# TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 4.473924e+06
# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations.
# TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 1.632792e+06
# HELP go_memstats_stack_inuse_bytes Number of bytes in use by the stack allocator.
# TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 655360
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 655360
# HELP go_memstats_sys_bytes Number of bytes obtained from system.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 7.4401032e+07
# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge
go_threads 13
# HELP node_boot_time_seconds Unix time of last boot, including microseconds.
# TYPE node_boot_time_seconds gauge
node_boot_time_seconds 1.627973899643539e+09
# HELP node_cpu_seconds_total Seconds the CPUs spent in each mode.
# TYPE node_cpu_seconds_total counter
node_cpu_seconds_total{cpu="0",mode="idle"} 78815.94
node_cpu_seconds_total{cpu="0",mode="nice"} 0
node_cpu_seconds_total{cpu="0",mode="system"} 9706.08
node_cpu_seconds_total{cpu="0",mode="user"} 17779.6
node_cpu_seconds_total{cpu="1",mode="idle"} 104277.61
node_cpu_seconds_total{cpu="1",mode="nice"} 0
node_cpu_seconds_total{cpu="1",mode="system"} 1001.93
node_cpu_seconds_total{cpu="1",mode="user"} 1011.87
node_cpu_seconds_total{cpu="10",mode="idle"} 95181.94
node_cpu_seconds_total{cpu="10",mode="nice"} 0
node_cpu_seconds_total{cpu="10",mode="system"} 3570.19
node_cpu_seconds_total{cpu="10",mode="user"} 7539.7
node_cpu_seconds_total{cpu="11",mode="idle"} 104624.99
node_cpu_seconds_total{cpu="11",mode="nice"} 0
node_cpu_seconds_total{cpu="11",mode="system"} 616.9
node_cpu_seconds_total{cpu="11",mode="user"} 1049.43
node_cpu_seconds_total{cpu="2",mode="idle"} 85543.83
node_cpu_seconds_total{cpu="2",mode="nice"} 0
node_cpu_seconds_total{cpu="2",mode="system"} 7439.83
node_cpu_seconds_total{cpu="2",mode="user"} 13308.29
node_cpu_seconds_total{cpu="3",mode="idle"} 104618.23
node_cpu_seconds_total{cpu="3",mode="nice"} 0
node_cpu_seconds_total{cpu="3",mode="system"} 724.34
node_cpu_seconds_total{cpu="3",mode="user"} 948.83
node_cpu_seconds_total{cpu="4",mode="idle"} 89505.26
node_cpu_seconds_total{cpu="4",mode="nice"} 0
node_cpu_seconds_total{cpu="4",mode="system"} 5812.29
node_cpu_seconds_total{cpu="4",mode="user"} 10974.36
node_cpu_seconds_total{cpu="5",mode="idle"} 104614.55
node_cpu_seconds_total{cpu="5",mode="nice"} 0
node_cpu_seconds_total{cpu="5",mode="system"} 695
node_cpu_seconds_total{cpu="5",mode="user"} 981.82
node_cpu_seconds_total{cpu="6",mode="idle"} 91356.45
node_cpu_seconds_total{cpu="6",mode="nice"} 0
node_cpu_seconds_total{cpu="6",mode="system"} 5106.29
node_cpu_seconds_total{cpu="6",mode="user"} 9829.14
node_cpu_seconds_total{cpu="7",mode="idle"} 104624.68
node_cpu_seconds_total{cpu="7",mode="nice"} 0
node_cpu_seconds_total{cpu="7",mode="system"} 660.45
node_cpu_seconds_total{cpu="7",mode="user"} 1006.22
node_cpu_seconds_total{cpu="8",mode="idle"} 92783.58
node_cpu_seconds_total{cpu="8",mode="nice"} 0
node_cpu_seconds_total{cpu="8",mode="system"} 4514.87
node_cpu_seconds_total{cpu="8",mode="user"} 8993.41
node_cpu_seconds_total{cpu="9",mode="idle"} 104642.2
node_cpu_seconds_total{cpu="9",mode="nice"} 0
node_cpu_seconds_total{cpu="9",mode="system"} 630.01
node_cpu_seconds_total{cpu="9",mode="user"} 1019.12
# HELP node_disk_read_bytes_total The total number of bytes read successfully.
# TYPE node_disk_read_bytes_total counter
node_disk_read_bytes_total{device="disk0"} 1.1184678912e+11
# HELP node_disk_read_errors_total The total number of read errors.
# TYPE node_disk_read_errors_total counter
node_disk_read_errors_total{device="disk0"} 0
# HELP node_disk_read_retries_total The total number of read retries.
# TYPE node_disk_read_retries_total counter
node_disk_read_retries_total{device="disk0"} 0
# HELP node_disk_read_sectors_total The total number of sectors read successfully.
# TYPE node_disk_read_sectors_total counter
node_disk_read_sectors_total{device="disk0"} 1071.09423828125
# HELP node_disk_read_time_seconds_total The total number of seconds spent by all reads.
# TYPE node_disk_read_time_seconds_total counter
node_disk_read_time_seconds_total{device="disk0"} 3123.997486813
# HELP node_disk_reads_completed_total The total number of reads completed successfully.
# TYPE node_disk_reads_completed_total counter
node_disk_reads_completed_total{device="disk0"} 4.387202e+06
# HELP node_disk_write_errors_total The total number of write errors.
# TYPE node_disk_write_errors_total counter
node_disk_write_errors_total{device="disk0"} 0
# HELP node_disk_write_retries_total The total number of write retries.
# TYPE node_disk_write_retries_total counter
node_disk_write_retries_total{device="disk0"} 0
# HELP node_disk_write_time_seconds_total This is the total number of seconds spent by all writes.
# TYPE node_disk_write_time_seconds_total counter
node_disk_write_time_seconds_total{device="disk0"} 769.724743697
# HELP node_disk_writes_completed_total The total number of writes completed successfully.
# TYPE node_disk_writes_completed_total counter
node_disk_writes_completed_total{device="disk0"} 2.666872e+06
# HELP node_disk_written_bytes_total The total number of bytes written successfully.
# TYPE node_disk_written_bytes_total counter
node_disk_written_bytes_total{device="disk0"} 9.9114201088e+10
# HELP node_disk_written_sectors_total The total number of sectors written successfully.
# TYPE node_disk_written_sectors_total counter
node_disk_written_sectors_total{device="disk0"} 651.091796875
# HELP node_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which node_exporter was built.
# TYPE node_exporter_build_info gauge
node_exporter_build_info{branch="",goversion="go1.16.5",revision="",version=""} 1
# HELP node_filesystem_avail_bytes Filesystem space available to non-root users in bytes.
# TYPE node_filesystem_avail_bytes gauge
node_filesystem_avail_bytes{device="/dev/disk1s1",fstype="apfs",mountpoint="/System/Volumes/Data"} 4.25092554752e+11
node_filesystem_avail_bytes{device="/dev/disk1s2",fstype="apfs",mountpoint="/System/Volumes/Preboot"} 4.25351725056e+11
node_filesystem_avail_bytes{device="/dev/disk1s4",fstype="apfs",mountpoint="/System/Volumes/VM"} 4.25351725056e+11
node_filesystem_avail_bytes{device="/dev/disk1s5s1",fstype="apfs",mountpoint="/"} 4.25092554752e+11
node_filesystem_avail_bytes{device="/dev/disk1s6",fstype="apfs",mountpoint="/System/Volumes/Update"} 4.25351725056e+11
node_filesystem_avail_bytes{device="map auto_home",fstype="autofs",mountpoint="/System/Volumes/Data/home"} 0
# HELP node_filesystem_device_error Whether an error occurred while getting statistics for the given device.
# TYPE node_filesystem_device_error gauge
node_filesystem_device_error{device="/dev/disk1s1",fstype="apfs",mountpoint="/System/Volumes/Data"} 0
node_filesystem_device_error{device="/dev/disk1s2",fstype="apfs",mountpoint="/System/Volumes/Preboot"} 0
node_filesystem_device_error{device="/dev/disk1s4",fstype="apfs",mountpoint="/System/Volumes/VM"} 0
node_filesystem_device_error{device="/dev/disk1s5s1",fstype="apfs",mountpoint="/"} 0
node_filesystem_device_error{device="/dev/disk1s6",fstype="apfs",mountpoint="/System/Volumes/Update"} 0
node_filesystem_device_error{device="map auto_home",fstype="autofs",mountpoint="/System/Volumes/Data/home"} 0
# HELP node_filesystem_files Filesystem total file nodes.
# TYPE node_filesystem_files gauge
node_filesystem_files{device="/dev/disk1s1",fstype="apfs",mountpoint="/System/Volumes/Data"} 4.88245288e+09
node_filesystem_files{device="/dev/disk1s2",fstype="apfs",mountpoint="/System/Volumes/Preboot"} 4.88245288e+09
node_filesystem_files{device="/dev/disk1s4",fstype="apfs",mountpoint="/System/Volumes/VM"} 4.88245288e+09
node_filesystem_files{device="/dev/disk1s5s1",fstype="apfs",mountpoint="/"} 4.88245288e+09
node_filesystem_files{device="/dev/disk1s6",fstype="apfs",mountpoint="/System/Volumes/Update"} 4.88245288e+09
node_filesystem_files{device="map auto_home",fstype="autofs",mountpoint="/System/Volumes/Data/home"} 0
# HELP node_filesystem_files_free Filesystem total free file nodes.
# TYPE node_filesystem_files_free gauge
node_filesystem_files_free{device="/dev/disk1s1",fstype="apfs",mountpoint="/System/Volumes/Data"} 4.881688286e+09
node_filesystem_files_free{device="/dev/disk1s2",fstype="apfs",mountpoint="/System/Volumes/Preboot"} 4.882451674e+09
node_filesystem_files_free{device="/dev/disk1s4",fstype="apfs",mountpoint="/System/Volumes/VM"} 4.882452878e+09
node_filesystem_files_free{device="/dev/disk1s5s1",fstype="apfs",mountpoint="/"} 4.881899123e+09
node_filesystem_files_free{device="/dev/disk1s6",fstype="apfs",mountpoint="/System/Volumes/Update"} 4.882452863e+09
node_filesystem_files_free{device="map auto_home",fstype="autofs",mountpoint="/System/Volumes/Data/home"} 0
# HELP node_filesystem_free_bytes Filesystem free space in bytes.
# TYPE node_filesystem_free_bytes gauge
node_filesystem_free_bytes{device="/dev/disk1s1",fstype="apfs",mountpoint="/System/Volumes/Data"} 4.4365203456e+11
node_filesystem_free_bytes{device="/dev/disk1s2",fstype="apfs",mountpoint="/System/Volumes/Preboot"} 4.99648794624e+11
node_filesystem_free_bytes{device="/dev/disk1s4",fstype="apfs",mountpoint="/System/Volumes/VM"} 4.97815670784e+11
node_filesystem_free_bytes{device="/dev/disk1s5s1",fstype="apfs",mountpoint="/"} 4.84650889216e+11
node_filesystem_free_bytes{device="/dev/disk1s6",fstype="apfs",mountpoint="/System/Volumes/Update"} 4.99962273792e+11
node_filesystem_free_bytes{device="map auto_home",fstype="autofs",mountpoint="/System/Volumes/Data/home"} 0
# HELP node_filesystem_readonly Filesystem read-only status.
# TYPE node_filesystem_readonly gauge
node_filesystem_readonly{device="/dev/disk1s1",fstype="apfs",mountpoint="/System/Volumes/Data"} 0
node_filesystem_readonly{device="/dev/disk1s2",fstype="apfs",mountpoint="/System/Volumes/Preboot"} 0
node_filesystem_readonly{device="/dev/disk1s4",fstype="apfs",mountpoint="/System/Volumes/VM"} 0
node_filesystem_readonly{device="/dev/disk1s5s1",fstype="apfs",mountpoint="/"} 1
node_filesystem_readonly{device="/dev/disk1s6",fstype="apfs",mountpoint="/System/Volumes/Update"} 0
node_filesystem_readonly{device="map auto_home",fstype="autofs",mountpoint="/System/Volumes/Data/home"} 0
# HELP node_filesystem_size_bytes Filesystem size in bytes.
# TYPE node_filesystem_size_bytes gauge
node_filesystem_size_bytes{device="/dev/disk1s1",fstype="apfs",mountpoint="/System/Volumes/Data"} 4.99963174912e+11
node_filesystem_size_bytes{device="/dev/disk1s2",fstype="apfs",mountpoint="/System/Volumes/Preboot"} 4.99963174912e+11
node_filesystem_size_bytes{device="/dev/disk1s4",fstype="apfs",mountpoint="/System/Volumes/VM"} 4.99963174912e+11
node_filesystem_size_bytes{device="/dev/disk1s5s1",fstype="apfs",mountpoint="/"} 4.99963174912e+11
node_filesystem_size_bytes{device="/dev/disk1s6",fstype="apfs",mountpoint="/System/Volumes/Update"} 4.99963174912e+11
node_filesystem_size_bytes{device="map auto_home",fstype="autofs",mountpoint="/System/Volumes/Data/home"} 0
# HELP node_load1 1m load average.
# TYPE node_load1 gauge
node_load1 2.5048828125
# HELP node_load15 15m load average.
# TYPE node_load15 gauge
node_load15 2.8779296875
# HELP node_load5 5m load average.
# TYPE node_load5 gauge
node_load5 2.86328125
# HELP node_memory_active_bytes Memory information field active_bytes.
# TYPE node_memory_active_bytes gauge
node_memory_active_bytes 5.238034432e+09
# HELP node_memory_compressed_bytes Memory information field compressed_bytes.
# TYPE node_memory_compressed_bytes gauge
node_memory_compressed_bytes 2.478948352e+09
# HELP node_memory_free_bytes Memory information field free_bytes.
# TYPE node_memory_free_bytes gauge
node_memory_free_bytes 5.01157888e+08
# HELP node_memory_inactive_bytes Memory information field inactive_bytes.
# TYPE node_memory_inactive_bytes gauge
node_memory_inactive_bytes 5.574938624e+09
# HELP node_memory_swap_total_bytes Memory information field swap_total_bytes.
# TYPE node_memory_swap_total_bytes gauge
node_memory_swap_total_bytes 2.147483648e+09
# HELP node_memory_swap_used_bytes Memory information field swap_used_bytes.
# TYPE node_memory_swap_used_bytes gauge
node_memory_swap_used_bytes 1.095499776e+09
# HELP node_memory_swapped_in_bytes_total Memory information field swapped_in_bytes_total.
# TYPE node_memory_swapped_in_bytes_total counter
node_memory_swapped_in_bytes_total 5.3633826816e+10
# HELP node_memory_swapped_out_bytes_total Memory information field swapped_out_bytes_total.
# TYPE node_memory_swapped_out_bytes_total counter
node_memory_swapped_out_bytes_total 4.1562112e+08
# HELP node_memory_total_bytes Memory information field total_bytes.
# TYPE node_memory_total_bytes gauge
node_memory_total_bytes 1.7179869184e+10
# HELP node_memory_wired_bytes Memory information field wired_bytes.
# TYPE node_memory_wired_bytes gauge
node_memory_wired_bytes 3.3830912e+09
# HELP node_network_receive_bytes_total Network device statistic receive_bytes.
# TYPE node_network_receive_bytes_total counter
node_network_receive_bytes_total{device="ap1"} 0
node_network_receive_bytes_total{device="awdl0"} 1.59578112e+08
node_network_receive_bytes_total{device="bridge0"} 0
node_network_receive_bytes_total{device="en0"} 3.019623424e+09
node_network_receive_bytes_total{device="en1"} 0
node_network_receive_bytes_total{device="en2"} 0
node_network_receive_bytes_total{device="en3"} 0
node_network_receive_bytes_total{device="en4"} 0
node_network_receive_bytes_total{device="en5"} 4.732928e+06
node_network_receive_bytes_total{device="gif0"} 0
node_network_receive_bytes_total{device="llw0"} 0
node_network_receive_bytes_total{device="lo0"} 2.22496768e+08
node_network_receive_bytes_total{device="stf0"} 0
node_network_receive_bytes_total{device="utun0"} 0
node_network_receive_bytes_total{device="utun1"} 0
node_network_receive_bytes_total{device="utun2"} 1024
node_network_receive_bytes_total{device="utun3"} 0
# HELP node_network_receive_errs_total Network device statistic receive_errs.
# TYPE node_network_receive_errs_total counter
node_network_receive_errs_total{device="ap1"} 0
node_network_receive_errs_total{device="awdl0"} 0
node_network_receive_errs_total{device="bridge0"} 0
node_network_receive_errs_total{device="en0"} 0
node_network_receive_errs_total{device="en1"} 0
node_network_receive_errs_total{device="en2"} 0
node_network_receive_errs_total{device="en3"} 0
node_network_receive_errs_total{device="en4"} 0
node_network_receive_errs_total{device="en5"} 0
node_network_receive_errs_total{device="gif0"} 0
node_network_receive_errs_total{device="llw0"} 0
node_network_receive_errs_total{device="lo0"} 0
node_network_receive_errs_total{device="stf0"} 0
node_network_receive_errs_total{device="utun0"} 0
node_network_receive_errs_total{device="utun1"} 0
node_network_receive_errs_total{device="utun2"} 0
node_network_receive_errs_total{device="utun3"} 0
# HELP node_network_receive_multicast_total Network device statistic receive_multicast.
# TYPE node_network_receive_multicast_total counter
node_network_receive_multicast_total{device="ap1"} 0
node_network_receive_multicast_total{device="awdl0"} 739225
node_network_receive_multicast_total{device="bridge0"} 0
node_network_receive_multicast_total{device="en0"} 61380
node_network_receive_multicast_total{device="en1"} 0
node_network_receive_multicast_total{device="en2"} 0
node_network_receive_multicast_total{device="en3"} 0
node_network_receive_multicast_total{device="en4"} 0
node_network_receive_multicast_total{device="en5"} 0
node_network_receive_multicast_total{device="gif0"} 0
node_network_receive_multicast_total{device="llw0"} 0
node_network_receive_multicast_total{device="lo0"} 59874
node_network_receive_multicast_total{device="stf0"} 0
node_network_receive_multicast_total{device="utun0"} 0
node_network_receive_multicast_total{device="utun1"} 0
node_network_receive_multicast_total{device="utun2"} 0
node_network_receive_multicast_total{device="utun3"} 0
# HELP node_network_receive_packets_total Network device statistic receive_packets.
# TYPE node_network_receive_packets_total counter
node_network_receive_packets_total{device="ap1"} 0
node_network_receive_packets_total{device="awdl0"} 740034
node_network_receive_packets_total{device="bridge0"} 0
node_network_receive_packets_total{device="en0"} 3.625992e+06
node_network_receive_packets_total{device="en1"} 0
node_network_receive_packets_total{device="en2"} 0
node_network_receive_packets_total{device="en3"} 0
node_network_receive_packets_total{device="en4"} 0
node_network_receive_packets_total{device="en5"} 39519
node_network_receive_packets_total{device="gif0"} 0
node_network_receive_packets_total{device="llw0"} 0
node_network_receive_packets_total{device="lo0"} 817644
node_network_receive_packets_total{device="stf0"} 0
node_network_receive_packets_total{device="utun0"} 0
node_network_receive_packets_total{device="utun1"} 0
node_network_receive_packets_total{device="utun2"} 26
node_network_receive_packets_total{device="utun3"} 0
# HELP node_network_transmit_bytes_total Network device statistic transmit_bytes.
# TYPE node_network_transmit_bytes_total counter
node_network_transmit_bytes_total{device="ap1"} 0
node_network_transmit_bytes_total{device="awdl0"} 3.00544e+06
node_network_transmit_bytes_total{device="bridge0"} 0
node_network_transmit_bytes_total{device="en0"} 2.886089728e+09
node_network_transmit_bytes_total{device="en1"} 0
node_network_transmit_bytes_total{device="en2"} 0
node_network_transmit_bytes_total{device="en3"} 0
node_network_transmit_bytes_total{device="en4"} 0
node_network_transmit_bytes_total{device="en5"} 4.0312832e+07
node_network_transmit_bytes_total{device="gif0"} 0
node_network_transmit_bytes_total{device="llw0"} 0
node_network_transmit_bytes_total{device="lo0"} 2.22496768e+08
node_network_transmit_bytes_total{device="stf0"} 0
node_network_transmit_bytes_total{device="utun0"} 8192
node_network_transmit_bytes_total{device="utun1"} 8192
node_network_transmit_bytes_total{device="utun2"} 1.0496e+06
node_network_transmit_bytes_total{device="utun3"} 8192
# HELP node_network_transmit_errs_total Network device statistic transmit_errs.
# TYPE node_network_transmit_errs_total counter
node_network_transmit_errs_total{device="ap1"} 0
node_network_transmit_errs_total{device="awdl0"} 0
node_network_transmit_errs_total{device="bridge0"} 0
node_network_transmit_errs_total{device="en0"} 18460
node_network_transmit_errs_total{device="en1"} 0
node_network_transmit_errs_total{device="en2"} 0
node_network_transmit_errs_total{device="en3"} 0
node_network_transmit_errs_total{device="en4"} 0
node_network_transmit_errs_total{device="en5"} 6021
node_network_transmit_errs_total{device="gif0"} 0
node_network_transmit_errs_total{device="llw0"} 0
node_network_transmit_errs_total{device="lo0"} 0
node_network_transmit_errs_total{device="stf0"} 0
node_network_transmit_errs_total{device="utun0"} 0
node_network_transmit_errs_total{device="utun1"} 0
node_network_transmit_errs_total{device="utun2"} 0
node_network_transmit_errs_total{device="utun3"} 0
# HELP node_network_transmit_multicast_total Network device statistic transmit_multicast.
# TYPE node_network_transmit_multicast_total counter
node_network_transmit_multicast_total{device="ap1"} 0
node_network_transmit_multicast_total{device="awdl0"} 0
node_network_transmit_multicast_total{device="bridge0"} 0
node_network_transmit_multicast_total{device="en0"} 0
node_network_transmit_multicast_total{device="en1"} 0
node_network_transmit_multicast_total{device="en2"} 0
node_network_transmit_multicast_total{device="en3"} 0
node_network_transmit_multicast_total{device="en4"} 0
node_network_transmit_multicast_total{device="en5"} 0
node_network_transmit_multicast_total{device="gif0"} 0
node_network_transmit_multicast_total{device="llw0"} 0
node_network_transmit_multicast_total{device="lo0"} 0
node_network_transmit_multicast_total{device="stf0"} 0
node_network_transmit_multicast_total{device="utun0"} 0
node_network_transmit_multicast_total{device="utun1"} 0
node_network_transmit_multicast_total{device="utun2"} 0
node_network_transmit_multicast_total{device="utun3"} 0
# HELP node_network_transmit_packets_total Network device statistic transmit_packets.
# TYPE node_network_transmit_packets_total counter
node_network_transmit_packets_total{device="ap1"} 0
node_network_transmit_packets_total{device="awdl0"} 9965
node_network_transmit_packets_total{device="bridge0"} 0
node_network_transmit_packets_total{device="en0"} 5.698822e+06
node_network_transmit_packets_total{device="en1"} 0
node_network_transmit_packets_total{device="en2"} 0
node_network_transmit_packets_total{device="en3"} 0
node_network_transmit_packets_total{device="en4"} 0
node_network_transmit_packets_total{device="en5"} 41456
node_network_transmit_packets_total{device="gif0"} 0
node_network_transmit_packets_total{device="llw0"} 0
node_network_transmit_packets_total{device="lo0"} 817644
node_network_transmit_packets_total{device="stf0"} 0
node_network_transmit_packets_total{device="utun0"} 47
node_network_transmit_packets_total{device="utun1"} 47
node_network_transmit_packets_total{device="utun2"} 12081
node_network_transmit_packets_total{device="utun3"} 47
# HELP node_scrape_collector_duration_seconds node_exporter: Duration of a collector scrape.
# TYPE node_scrape_collector_duration_seconds gauge
node_scrape_collector_duration_seconds{collector="boottime"} 0.000204068
node_scrape_collector_duration_seconds{collector="cpu"} 0.00014535
node_scrape_collector_duration_seconds{collector="diskstats"} 0.000605205
node_scrape_collector_duration_seconds{collector="filesystem"} 9.2548e-05
node_scrape_collector_duration_seconds{collector="loadavg"} 3.6591e-05
node_scrape_collector_duration_seconds{collector="meminfo"} 0.000137689
node_scrape_collector_duration_seconds{collector="netdev"} 0.000754745
node_scrape_collector_duration_seconds{collector="textfile"} 0.000204323
node_scrape_collector_duration_seconds{collector="time"} 2.1172e-05
node_scrape_collector_duration_seconds{collector="uname"} 6.1507e-05
# HELP node_scrape_collector_success node_exporter: Whether a collector succeeded.
# TYPE node_scrape_collector_success gauge
node_scrape_collector_success{collector="boottime"} 1
node_scrape_collector_success{collector="cpu"} 1
node_scrape_collector_success{collector="diskstats"} 1
node_scrape_collector_success{collector="filesystem"} 1
node_scrape_collector_success{collector="loadavg"} 1
node_scrape_collector_success{collector="meminfo"} 1
node_scrape_collector_success{collector="netdev"} 1
node_scrape_collector_success{collector="textfile"} 1
node_scrape_collector_success{collector="time"} 1
node_scrape_collector_success{collector="uname"} 1
# HELP node_textfile_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE node_textfile_scrape_error gauge
node_textfile_scrape_error 0
# HELP node_time_seconds System time in seconds since epoch (1970).
# TYPE node_time_seconds gauge
node_time_seconds 1.6281710050197089e+09
# HELP node_time_zone_offset_seconds System time zone offset in seconds.
# TYPE node_time_zone_offset_seconds gauge
node_time_zone_offset_seconds{time_zone="CST"} 28800
# HELP node_uname_info Labeled system information as provided by the uname system call.
# TYPE node_uname_info gauge
node_uname_info{domainname="local",machine="x86_64",nodename="MacBook-Pro-16",release="20.5.0",sysname="Darwin",version="Darwin Kernel Version 20.5.0: Sat May  8 05:10:33 PDT 2021; root:xnu-7195.121.3~9/RELEASE_X86_64"} 1
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 2
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
`
