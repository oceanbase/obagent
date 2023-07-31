package log_query

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/config/mgragent"
	"github.com/oceanbase/obagent/lib/log_analyzer"
)

func TestLogQuerier_queryLogByLine(t *testing.T) {
	fileInfo := &FileInfo{
		FileName:   "test.log",
		FileId:     1,
		FileOffset: 0,
	}

	now := time.Now()
	observerLogAnalyzer := log_analyzer.NewObLogLightAnalyzer(fileInfo.FileName)
	logQuery := &LogQuery{
		excludeKeywordRegexps: make([]*regexp.Regexp, 0),
		queryLogParams: &QueryLogRequest{
			StartTime:           now,
			EndTime:             now.AddDate(0, 0, 1),
			LogType:             "observer",
			Keyword:             make([]string, 0),
			ExcludeKeyword:      make([]string, 0),
			LogLevel:            []string{"INFO"},
			ReqId:               "1",
			LastQueryFileId:     0,
			LastQueryFileOffset: 0,
			Limit:               20,
		},
		logEntryChan: make(chan LogEntry, 1),
		count:        0,
	}

	logQuerier := NewLogQuerier(&mgragent.LogQueryConfig{ErrCountLimit: 100})
	logLine := `[2021-12-06 10:03:00.086260] INFO  [SERVER] ob_query_retry_ctrl.cpp:557  test line 1
[2021-12-06 10:03:00.086270] WARN  [SERVER] response_result (ob_sync_plan_driver.cpp:74) [1931][1750] test line 2
[2021-12-06 10:03:00.086298] WARN  [SERVER] response_result (ob_sync_plan_driver.cpp:81) [1931][1750] test line 3
[2021-12-06 10:03:00.086317] INFO  [SERVER] obmp_base.cpp:1230 [1931][1750][YB420BA64D8A-0005D10D459FD4EF] test line 4`
	strReader := strings.NewReader(logLine)

	type args struct {
		fileInfo  *FileInfo
		logQuery  *LogQuery
		strReader io.Reader
	}
	tests := []struct {
		name string
		args args
		want *Position
	}{
		{
			name: "根据 logQuery 查询匹配的日志",
			args: args{
				fileInfo:  fileInfo,
				logQuery:  logQuery,
				strReader: strReader,
			},
			want: &Position{
				FileId:     1,
				FileOffset: int64(len(logLine)),
			},
		},
	}

	for _, tt := range tests {
		Convey(tt.name, t, func() {
			lastPos, err := logQuerier.queryLogByLine(context.Background(), fileInfo, strReader, logQuery, observerLogAnalyzer)

			So(err, ShouldBeNil)
			So(lastPos.FileId, ShouldEqual, tt.want.FileId)
			So(lastPos.FileOffset, ShouldEqual, tt.want.FileOffset)
		})
	}
}

func prepareTestDirTree(tree string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v\n", err)
	}

	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		return "", fmt.Errorf("error evaling temp directory: %v\n", err)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, tree), 0755)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return filepath.Join(tmpDir, tree), nil
}

func TestLogQuerier_getMatchedFiles(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2022, 3, 31, 13, 33, 3, 3, time.Local)
	logFileName := "observer.log.20220331005827"
	os.Create(filepath.Join(tmpDir, logFileName))

	defer os.RemoveAll(tmpDir)
	logQuerier := &LogQuerier{}
	logQuery := &LogQuery{
		conf: mgragent.LogQueryConfig{
			ErrCountLimit:   10,
			QueryTimeout:    1000,
			DownloadTimeout: 1000,
			LogTypeQueryConfigs: []mgragent.LogTypeQueryConfig{
				{
					LogType:              "observer",
					IsOverrideByPriority: true,
					LogLevelAndFilePatterns: []mgragent.LogLevelAndFilePattern{
						{
							LogLevel:          "ERROR",
							Dir:               tmpDir,
							FilePatterns:      []string{"observer.log.wf*"},
							LogParserCategory: "ob_light",
						}, {
							LogLevel:          "WARN",
							Dir:               tmpDir,
							FilePatterns:      []string{"observer.log.wf*"},
							LogParserCategory: "ob_light",
						}, {
							LogLevel:          "INFO",
							Dir:               tmpDir,
							FilePatterns:      []string{"observer.log*"},
							LogParserCategory: "ob_light",
						}, {
							LogLevel:          "DEBUG",
							Dir:               tmpDir,
							FilePatterns:      []string{"observer.log*"},
							LogParserCategory: "ob_light",
						},
					},
				}, {
					LogType:              "election",
					IsOverrideByPriority: true,
					LogLevelAndFilePatterns: []mgragent.LogLevelAndFilePattern{
						{
							LogLevel:          "ERROR",
							Dir:               tmpDir,
							FilePatterns:      []string{"election.log.wf*"},
							LogParserCategory: "ob_light",
						}, {
							LogLevel:          "WARN",
							Dir:               tmpDir,
							FilePatterns:      []string{"election.log.wf*"},
							LogParserCategory: "ob_light",
						}, {
							LogLevel:          "INFO",
							Dir:               tmpDir,
							FilePatterns:      []string{"election.log*"},
							LogParserCategory: "ob_light",
						}, {
							LogLevel:          "DEBUG",
							Dir:               tmpDir,
							FilePatterns:      []string{"election.log*"},
							LogParserCategory: "ob_light",
						},
					},
				},
			},
		},
		queryLogParams: &QueryLogRequest{
			StartTime:           now.AddDate(0, 0, -1),
			EndTime:             now.AddDate(0, 0, 1),
			LogType:             "observer",
			Keyword:             []string{"test"},
			LogLevel:            []string{"INFO"},
			LastQueryFileId:     0,
			LastQueryFileOffset: 0,
			Limit:               10,
		},
	}
	fileDetailInfos, err := logQuerier.getMatchedFiles(context.Background(), logQuery)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(fileDetailInfos))
	if len(fileDetailInfos) == 1 {
		fileDetailInfo1 := fileDetailInfos[0]
		assert.Equal(t, tmpDir, fileDetailInfo1.Dir)
		assert.Equal(t, logFileName, fileDetailInfo1.FileInfo.Name())
		assert.NotNil(t, fileDetailInfo1.LogAnalyzer)
	}
}

func TestLogQuerier_locateStartPosition(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	logFileName := "observer.log"
	file, err := os.OpenFile(filepath.Join(tmpDir, logFileName), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString(`[2022-03-31 17:52:30.796493] INFO  [CLOG] ob_log_callback_engine.cpp:84 [1667][0][Y0-0000000000000000] [lt=17] [dc=0] callback queue task number(clog=-7, ret=0)
[2022-03-31 17:52:40.796642] INFO  [CLOG] ob_log_flush_task.cpp:150 [1700][1078][YB420BA64D8A-0005DB42F83CA516] [lt=18] [dc=0] clog flush cb cost time(flush_cb_cnt=61, flush_cb_cost_time=2115, avg time=34)
[2022-03-31 17:52:41.800244] INFO  [SHARE] ob_bg_thread_monitor.cpp:323 [2263][2043][Y0-0000000000000000] [lt=18] [dc=0] current monitor number(seq_=0)
[2022-03-31 17:52:42.821926] INFO  [COMMON] ob_kvcache_store.cpp:811 [1389][464][Y0-0000000000000000] [lt=15] [dc=0] Wash compute wash size(sys_total_wash_size=-99128180736, global_cache_size=9640219456, tenant_max_wash_size=0, tenant_min_wash_size=0, tenant_ids_=[1, 500, 1001, 1002], sys_cache_reserve_size=419430400, tg=time guard 'compute_tenant_wash_size' cost too much time, used=160, time_dist: 112,4,1,12,3,0)
[2022-03-31 17:52:43.822301] INFO  [COMMON] ob_kvcache_store.cpp:318 [1389][464][Y0-0000000000000000] [lt=39] [dc=0] Wash time detail, (refresh_score_time=843, compute_wash_size_time=179, wash_sort_time=332, wash_time=2)
[2022-03-31 17:52:44.823868] INFO  [STORAGE] ob_partition_loop_worker.cpp:404 [2126][1922][Y0-0000000000000000] [lt=19] [dc=0] gene checkpoint(pkey={tid:1099511627961, partition_id:2, part_cnt:0}, state=6, last_checkpoint=1648720420756550, cur_checkpoint=1648720420756550, last_max_trans_version=1648455281795330, max_trans_version=1648455281795330)
[2022-03-31 17:52:49.835267] INFO  [SERVER] ob_inner_sql_connection.cpp:1336 [2055][1788][Y0-0000000000000000] [lt=18] [dc=0] execute write sql(ret=0, tenant_id=1, affected_rows=1, sql="     update __all_weak_read_service set min_version=1648720420707406, max_version=1648720420707406     where level_id = 0 and level_value = '' and min_version = 1648720420657154 and max_version = 1648720420657154 ")
[2022-03-31 17:52:51.844175] INFO  [COMMON] ob_kvcache_store.cpp:811 [1389][464][Y0-0000000000000000] [lt=13] [dc=0] Wash compute wash size(sys_total_wash_size=-99128180736, global_cache_size=9640219456, tenant_max_wash_size=0, tenant_min_wash_size=0, tenant_ids_=[1, 500, 1001, 1002], sys_cache_reserve_size=419430400, tg=time guard 'compute_tenant_wash_size' cost too much time, used=148, time_dist: 112,2,1,10,1,1)
[2022-03-31 17:52:56.844544] INFO  [COMMON] ob_kvcache_store.cpp:318 [1389][464][Y0-0000000000000000] [lt=37] [dc=0] Wash time detail, (refresh_score_time=768, compute_wash_size_time=170, wash_sort_time=328, wash_time=2)
[2022-03-31 17:52:59.877324] INFO  [STORAGE] ob_freeze_info_snapshot_mgr.cpp:982 [1644][970][Y0-0000000000000000] [lt=21] [dc=0] start reload freeze info and snapshots(is_remote_=true)
[2022-03-31 17:53:01.895471] INFO  [SERVER] ob_inner_sql_connection.cpp:1336 [2055][1788][Y0-0000000000000000] [lt=22] [dc=0] execute write sql(ret=0, tenant_id=1, affected_rows=1, sql="     update __all_weak_read_service set min_version=1648720420757606, max_version=1648720420757606     where level_id = 0 and level_value = '' and min_version = 1648720420707406 and max_version = 1648720420707406 ")
[2022-03-31 17:53:10.927528] INFO  [STORAGE] ob_partition_loop_worker.cpp:374 [2126][1922][Y0-0000000000000000] [lt=29] [dc=0] write checkpoint success(pkey={tid:1100611139454249, partition_id:0, part_cnt:0}, cur_checkpoint=1648720420811993)
[2022-03-31 17:53:12.939917] INFO  [CLOG.EXTLOG] ob_external_fetcher.cpp:276 [1609][900][Y0-0000000000000000] [lt=23] [dc=0] [FETCH_LOG_STREAM] Wash Stream: wash expired stream success(count=0, retired_arr=[])
[2022-03-31 17:53:14.947401] INFO  [LIB] ob_json.cpp:278 [1294][274][Y0-0000000000000000] [lt=16] [dc=0] invalid token type, maybe it is valid empty json type(cur_token_.type=93, ret=-5006)
[2022-03-31 17:53:16.947443] INFO  ob_config.cpp:956 [1294][274][Y0-0000000000000000] [lt=20] [dc=0] succ to format_option_str(src="ASYNC NET_TIMEOUT = 30000000", dest="ASYNC NET_TIMEOUT  =  30000000")
[2022-03-31 17:53:20.948222] INFO  [LIB] ob_json.cpp:278 [1294][274][Y0-0000000000000000] [lt=11] [dc=0] invalid token type, maybe it is valid empty json type(cur_token_.type=93, ret=-5006)
[2022-03-31 17:53:31.948249] INFO  ob_config.cpp:956 [1294][274][Y0-0000000000000000] [lt=15] [dc=0] succ to format_option_str(src="ASYNC NET_TIMEOUT = 30000000", dest="ASYNC NET_TIMEOUT  =  30000000")
[2022-03-31 17:53:40.948339] INFO  [SERVER] ob_remote_server_provider.cpp:208 [1294][274][Y0-0000000000000000] [lt=8] [dc=0] [remote_server_provider] refresh server list(ret=0, ret="OB_SUCCESS", all_server_count=0)`)

	defer os.RemoveAll(tmpDir)
	logQuerier := &LogQuerier{minPosGap: 0}
	queryTime := time.Date(2022, 3, 31, 17, 53, 16, 0, time.Local)
	offset, err := logQuerier.locateStartPosition(context.Background(), *file, log_analyzer.GetLogAnalyzer(log_analyzer.TypeObLight, logFileName), queryTime)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotZero(t, offset)
}
