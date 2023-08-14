/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package common

import (
	"fmt"
	"os"
	"testing"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/auth"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestGetTargetWithParameter(t *testing.T) {
	dbConfig := &DbConnectionConfig{
		Url:     "abc:xxx@tcp(1.1.1.1:2881)/ob?a=b",
		MaxOpen: 32,
		MaxIdle: 1,
	}

	target := dbConfig.Target()
	require.Equal(t, "(1.1.1.1:2881)/ob", target)
}

func TestGetTargetWithoutParameter(t *testing.T) {
	dbConfig := &DbConnectionConfig{
		Url:     "abc:xxx@tcp(1.1.1.1:2881)/ob",
		MaxOpen: 32,
		MaxIdle: 1,
	}

	target := dbConfig.Target()
	require.Equal(t, "(1.1.1.1:2881)/ob", target)
}

func getTestObserver() (*Observer, error) {
	dbConnectionConfig := &DbConnectionConfig{
		Url:     "user:pass@tcp(127.0.0.1:9878)/oceanbase?timeout=5s&interpolateParams=true",
		MaxOpen: 32,
		MaxIdle: 2,
	}
	ob := &Observer{
		DbConfig: dbConnectionConfig,
	}

	db, err := sqlx.Connect("mysql", ob.DbConfig.Url)
	ob.Db = db

	ob.MetaInfo = &ObserverMetaInfo{}
	ob.MetaInfo.Ip = "1.1.1.1"
	ob.MetaInfo.Port = 2881
	ob.RefreshTask()
	return ob, err
}

func TestGetObserverFirstTime(t *testing.T) {
	dbConnectionConfig := &DbConnectionConfig{
		Url:     "user:pass@tcp(127.0.0.1:9878)/oceanbase?timeout=5s",
		MaxOpen: 32,
		MaxIdle: 2,
	}
	_, err := GetObserver(dbConnectionConfig)
	fmt.Printf("got error: %v", err)
	require.True(t, err != nil)
}

func TestGetObserverWithDifferentUrl(t *testing.T) {

	dbConnectionConfig1 := &DbConnectionConfig{
		Url:     "user1:pass1@tcp(127.0.0.1:9878)/oceanbase?timeout=5s",
		MaxOpen: 32,
		MaxIdle: 2,
	}
	ob1, err := GetObserver(dbConnectionConfig1)
	require.True(t, err != nil)
	require.True(t, ob1 == nil)
}

func TestRefresh(t *testing.T) {
	ob, err := getTestObserver()
	require.True(t, err == nil)
	ob.RefreshTask()
	require.Equal(t, int64(0), ob.MetaInfo.Id)
}

func TestMain(m *testing.M) {
	s := setup()
	code := m.Run()
	s.Close()
	os.Exit(code)
}

func setup() *server.Server {
	driver := sqle.NewDefault()
	driver.AddDatabase(memory.NewDatabase("oceanbase"))

	config := server.Config{
		Protocol: "tcp",
		Address:  "0.0.0.0:9878",
		Version:  "2.2.77",
		Auth:     auth.NewNativeSingle("user", "pass", auth.AllPermissions),
	}

	s, err := server.NewDefaultServer(config, driver)
	if err != nil {
		panic(err)
	}

	go s.Start()
	return s
}
