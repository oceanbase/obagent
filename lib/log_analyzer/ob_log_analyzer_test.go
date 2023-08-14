/*
 * Copyright (c) 2023 OceanBase
 * OBAgent is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package log_analyzer

import (
	"bufio"
	"strings"
	"testing"
	"time"

	"github.com/oceanbase/obagent/monitor/message"
)

func sendLog(ch chan<- string, s string) {
	split := strings.Split(s, "\n")
	for _, line := range split {
		ch <- line
	}
	close(ch)
}

func drain(ch <-chan *message.Message) []*message.Message {
	var ret []*message.Message
	for s := range ch {
		ret = append(ret, s)
	}
	return ret
}

func checkTag(msg *message.Message, name, value string, t *testing.T) {
	if v, ok := msg.GetTag(name); !ok || v != value {
		t.Errorf("%s wrong. expected: %s, got: %s", name, value, v)
	}
}

func checkContent(msg *message.Message, value string, t *testing.T) {
	if c, ok := msg.GetField("content"); ok {
		if c.(string) != value {
			t.Error("content wrong")
		}
	} else {
		t.Error("no content")
	}
}

func TestParse1(t *testing.T) {
	in := make(chan string)
	out := make(chan *message.Message)
	go sendLog(in, `[2022-01-20 10:49:14.332262] INFO  [LIB] ob_json.cpp:278 [3451815][274][Y0-0000000000000000] [lt=14] [dc=0] invalid token type, maybe it is valid empty json type(cur_token_.type=93, ret=-5006)
[2022-01-20 10:49:14.332302] INFO  ob_config.cpp:956 [3451815][274][Y0-0000000000000000] [lt=21] [dc=0] succ to format_option_str(src="ASYNC NET_TIMEOUT = 30000000", dest="ASYNC NET_TIMEOUT  =  30000000")
[2022-01-20 10:49:14.332413] INFO  [SERVER] ob_remote_server_provider.cpp:208 [3451815][274][Y0-0000000000000000] [lt=8] [dc=0] [remote_server_provider] refresh server list(ret=0, ret="OB_SUCCESS", all_server_count=0)`)
	var msgs []*message.Message
	var err error
	analyzer := NewObLogAnalyzer("log1.log")
	go func() {
		err = ParseChan(analyzer, in, out)
	}()
	msgs = drain(out)
	if err != nil {
		t.Error("parse fail", err)
		return
	}
	if len(msgs) != 3 {
		t.Errorf("msg count wrong: %d", len(msgs))
	}
	checkTag(msgs[0], "module", "LIB", t)
	checkTag(msgs[0], "source", "ob_json.cpp", t)
	checkTag(msgs[0], "tid", "3451815", t)
	checkTag(msgs[0], "obLogTrace", "Y0-0000000000000000", t)
	checkTag(msgs[0], "level", "info", t)
	checkTag(msgs[0], "errCode", "5006", t)

	checkTag(msgs[2], "module", "SERVER", t)
	checkTag(msgs[2], "source", "ob_remote_server_provider.cpp", t)
	checkTag(msgs[2], "tid", "3451815", t)
	checkTag(msgs[2], "obLogTrace", "Y0-0000000000000000", t)
	checkTag(msgs[2], "level", "info", t)
}

func TestParseSingleLine(t *testing.T) {
	in := make(chan string)
	out := make(chan *message.Message)
	go sendLog(in, `[2022-01-20 10:49:14.332262] INFO  [LIB] ob_json.cpp:278 [3451815][274][Y0-0000000000000000] [lt=14] [dc=0] invalid token type, maybe it is valid empty json type(cur_token_.type=93, ret=-5006)`)
	var msgs []*message.Message
	var err error
	analyzer := NewObLogAnalyzer("log1.log")
	go func() {
		err = ParseChan(analyzer, in, out)
	}()
	msgs = drain(out)
	if err != nil {
		t.Error("parse fail", err)
		return
	}
	if len(msgs) != 1 {
		t.Errorf("msg count wrong: %d", len(msgs))
	}
	if msgs[0].GetName() != "log1.log" {
		t.Errorf("msg name wrong")
	}
	checkTag(msgs[0], "module", "LIB", t)
	checkTag(msgs[0], "source", "ob_json.cpp", t)
	checkTag(msgs[0], "tid", "3451815", t)
	checkTag(msgs[0], "obLogTrace", "Y0-0000000000000000", t)
	checkTag(msgs[0], "level", "info", t)
	checkTag(msgs[0], "errCode", "5006", t)
}

func TestParseSingleLineWithExtra(t *testing.T) {
	in := make(chan string)
	out := make(chan *message.Message)
	go sendLog(in, `[2022-01-20 10:44:04.983184] INFO  [CLOG] ob_log_state_driver_runnable.cpp:190 [3452682][1954][Y0-0000000000000000] [lt=26] [dc=0] ObLogStateDriverRunnable log delay histogram(histogram="Count: 0  Average: 0.0000  StdDev: 0.00
Min: 0.0000  Median: -nan  Max: 0.0000
------------------------------------------------------
")`)
	var msgs []*message.Message
	var err error
	analyzer := NewObLogAnalyzer("log1.log")
	go func() {
		err = ParseChan(analyzer, in, out)
	}()
	msgs = drain(out)
	if err != nil {
		t.Error("parse fail", err)
		return
	}
	if len(msgs) != 1 {
		t.Errorf("msg count wrong: %d", len(msgs))
	}
	if msgs[0].GetName() != "log1.log" {
		t.Errorf("msg name wrong")
	}
	checkTag(msgs[0], "module", "CLOG", t)
	checkTag(msgs[0], "source", "ob_log_state_driver_runnable.cpp", t)
	if v, ok := msgs[0].GetField("extra"); !ok || v != "Min: 0.0000  Median: -nan  Max: 0.0000\n------------------------------------------------------\n\")" {
		t.Error("extra wrong", v)
	}
}

func TestEmpty(t *testing.T) {
	in := make(chan string)
	out := make(chan *message.Message)

	var msgs []*message.Message
	var err error
	analyzer := NewObLogAnalyzer("log1.log")
	go func() {
		err = ParseChan(analyzer, in, out)
	}()
	close(in)
	msgs = drain(out)
	if err != nil {
		t.Error("parse fail", err)
		return
	}
	if len(msgs) != 0 {
		t.Errorf("msg count wrong: %d", len(msgs))
	}
}

func TestProxyLog(t *testing.T) {
	content := `[2022-03-23 02:34:25.994746] INFO  [PROXY.SM] ob_mysql_sm.cpp:438 [4135129][Y0-7F51F728E2E0] [lt=5] [dc=0] the request already in buffer, continue to handle it(buffer len=0, is_auth_rquest=true)
[2022-03-23 02:34:25.994808] WARN  [PROXY.SM] tunnel_handler_client (ob_mysql_sm.cpp:4150) [4135129][Y0-7F51F728E2E0] [lt=6] [dc=0] ObMysqlSM::tunnel_handler_client(event="VC_EVENT_EOS", sm_id=3671588)
[2022-03-23 02:34:25.994826] WARN  [PROXY.SM] set_client_abort (ob_mysql_sm.cpp:5675) [4135129][Y0-7F51F728E2E0] [lt=15] [dc=0] client will abort soon(sm_id=3671588, cs_id=130872, proxy_sessid=0, ss_id=0, server_sessid=0, client_ip={127.0.0.1:53581}, server_ip={*Not IP address [0]*:0}, cluster_name=, tenant_name=, user_name=, db=, event="VC_EVENT_EOS", request_cmd="Sleep", sql_cmd="Handshake", sql=COM_HANDSHAKE)
[2022-03-23 02:34:25.994860] INFO  [PROXY.CS] ob_mysql_client_session.cpp:83 [4135129][Y0-7F51F728E2E0] [lt=6] [dc=0] client session destroy(cs_id=130872, proxy_sessid=0, client_vc=NULL)
[2022-03-23 02:34:25.994884] INFO  [PROXY.SM] ob_mysql_sm.cpp:7000 [4135129][Y0-7F51F728E2E0] [lt=8] [dc=0] deallocating sm(sm_id=3671588)
[2022-05-12 15:05:23.535020] state_server_request_send (/ob_mysql_sm.cpp:4150) [130229][Y0-00007F48072C76A0-0-41] [lt=4] {"trace_id":"baa10601-41fd-bbce-aaba-812ecbde0500","name":"ObProxyServerRequestWrite","id":"90b4bdd7-c68f-0e45-82bc-812ecbde0500","start_ts":1652339123534978,"end_ts":1652339123535016,"parent_id":"399bcf38-b357-179f-b6ba-812ecbde0500","is_follow":false}`
	scanner := bufio.NewScanner(strings.NewReader(content))
	var msgs []*message.Message
	analyzer := NewObLogAnalyzer("log1.log")
	err := ParseScanner(analyzer, scanner, func(m *message.Message) bool { msgs = append(msgs, m); return true })
	if err != nil {
		t.Error("parse fail", err)
		return
	}
	if len(msgs) != 6 {
		t.Errorf("msg count wrong: %d", len(msgs))
	}
	checkTag(msgs[0], "obLogTrace", "Y0-7F51F728E2E0", t)
	checkTag(msgs[4], "module", "PROXY.SM", t)
	checkTag(msgs[4], "source", "ob_mysql_sm.cpp", t)
	checkTag(msgs[1], "func", "tunnel_handler_client", t)
}

func Test_parseNew(t *testing.T) {
	content := `[2022-05-10 19:00:58.231317] TestBody (test_trace.cpp:73) [93941][th1][T1000][Y0-0000000000000000-0-0] message content`
	analyzer := NewObLogAnalyzer("log1.log")
	msg, isNewLine := analyzer.ParseLine(content)
	if !isNewLine {
		t.Error("should be new line")
	}
	//msg.GetTag("")
	checkTag(msg, "tid", "93941", t)
	checkTag(msg, "obLogTrace", "Y0-0000000000000000-0-0", t)
	checkTag(msg, "tenant", "1000", t)
	checkTag(msg, "thread", "th1", t)
	checkTag(msg, "func", "TestBody", t)
	checkContent(msg, "message content", t)
	//
	content = `[2022-05-10 20:33:19.462086] INFO  [STORAGE.TRANS] print_stat_info (ob_keep_alive_ls_handler.cpp:130) [49476][WeakReadSvr][T1][Y0-0000000000000000-0-0] [Keep Alive Stat] LS Keep Alive Info(tenant_id=1, LS_ID={id:1}, Not_Master_Cnt=0, Near_To_GTS_Cnt=1, Other_Error_Cnt=0, Submit_Succ_Cnt=0, last_log_ts=0, last_lsn={val:18446744073709551615}, last_gts=1652185999374013306)`
	msg, _ = analyzer.ParseLine(content)
	checkTag(msg, "level", "info", t)
	checkTag(msg, "module", "STORAGE.TRANS", t)
	checkTag(msg, "module", "STORAGE.TRANS", t)
	checkTag(msg, "tid", "49476", t)
	checkTag(msg, "thread", "WeakReadSvr", t)
	checkTag(msg, "tenant", "1", t)

}

func Test_parseTimeFromFileName(t *testing.T) {
	now := time.Now()
	expected := time.Date(2022, time.January, 2, 11, 12, 13, 0, time.Local)
	parsed := ParseTimeFromFileName("file1.log.20220102111213", ".", obLogFileTimePattern, now)
	if expected != parsed {
		t.Error("parse file time wrong")
	}
	parsed = ParseTimeFromFileName("file1.log.1", ".", obLogFileTimePattern, now)
	if parsed != now {
		t.Error("should be default time")
	}
	parsed = ParseTimeFromFileName("file1xxxx", ".", obLogFileTimePattern, now)
	if parsed != now {
		t.Error("should be default time")
	}
}
