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

package obcommon

import (
	"os"
	"testing"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/auth"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
)

func TestMain(m *testing.M) {
	s := setup()
	code := m.Run()
	teardown(s)
	os.Exit(code)
}

func setup() *server.Server {
	driver := sqle.NewDefault()
	driver.AddDatabase(createTestDatabase())

	config := server.Config{
		Protocol: "tcp",
		Address:  "0.0.0.0:9878",
		Auth:     auth.NewNativeSingle("user", "pass", auth.AllPermissions),
	}

	s, err := server.NewDefaultServer(config, driver)
	if err != nil {
		panic(err)
	}

	go s.Start()
	return s
}

func teardown(s *server.Server) {
	s.Close()
}

func createTestDatabase() *memory.Database {
	const (
		oceanbase       = "oceanbase"
		allServer       = "__all_server"
		allUnit         = "__all_unit"
		allResourcePool = "__all_resource_pool"
	)

	// new database
	db := memory.NewDatabase(oceanbase)

	// mock table __all_server
	tableAllServer := memory.NewTable(allServer, sql.Schema{
		{Name: "id", Type: sql.Int64, Nullable: false, Source: allServer},
		{Name: "build_version", Type: sql.Text, Nullable: false, Source: allServer},
		{Name: "svr_ip", Type: sql.Text, Nullable: false, Source: allServer},
		{Name: "svr_port", Type: sql.Int64, Nullable: false, Source: allServer},
	})

	// mock table __all_unit
	tableAllUnit := memory.NewTable(allUnit, sql.Schema{
		{Name: "resource_pool_id", Type: sql.Int64, Nullable: false, Source: allUnit},
		{Name: "svr_ip", Type: sql.Text, Nullable: false, Source: allUnit},
		{Name: "svr_port", Type: sql.Int64, Nullable: false, Source: allUnit},
	})

	// mock table __all_resource_pool
	tableAllResourcePool := memory.NewTable(allResourcePool, sql.Schema{
		{Name: "resource_pool_id", Type: sql.Int64, Nullable: false, Source: allResourcePool},
		{Name: "tenant_id", Type: sql.Int64, Nullable: false, Source: allResourcePool},
	})

	// add table to db
	db.AddTable(allServer, tableAllServer)
	db.AddTable(allUnit, tableAllUnit)
	db.AddTable(allResourcePool, tableAllResourcePool)

	ctx := sql.NewEmptyContext()

	// write all server data
	allServerRows := []sql.Row{
		sql.NewRow(1, "2.2.0_20210716222653-e9b76c115d164f0158af3aae5be3854faf084012", "1.1.1.1", 2881),
	}
	for _, row := range allServerRows {
		tableAllServer.Insert(ctx, row)
	}

	// write all unit data
	allUnitRows := []sql.Row{
		sql.NewRow(1, "1.1.1.1", 2881),
		sql.NewRow(1001, "1.1.1.1", 2881),
	}
	for _, row := range allUnitRows {
		tableAllUnit.Insert(ctx, row)
	}

	// write all resource pool data
	allResourcePoolRows := []sql.Row{
		sql.NewRow(1, 1),
		sql.NewRow(1001, 1001),
	}
	for _, row := range allResourcePoolRows {
		tableAllResourcePool.Insert(ctx, row)
	}

	return db
}
