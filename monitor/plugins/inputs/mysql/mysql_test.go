package mysql

import (
	"fmt"
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
	driver.Catalog.MustRegister(
		sql.FunctionN{
			Name: "host_ip",
			Fn:   NewHostIp("127.0.0.1"),
		},
		sql.FunctionN{
			Name: "rpc_port",
			Fn:   NewPort("9878"),
		})

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

func createTestDatabase() *memory.Database {
	const (
		dbName    = "oceanbase"
		tableName = "test_metric"
	)

	db := memory.NewDatabase(dbName)
	table := memory.NewTable(tableName, sql.Schema{
		{Name: "t1", Type: sql.Text, Nullable: false, Source: tableName},
		{Name: "t2", Type: sql.Text, Nullable: false, Source: tableName},
		{Name: "m1", Type: sql.Float64, Nullable: false, Source: tableName},
		{Name: "m2", Type: sql.Int64, Nullable: false, Source: tableName},
	})

	db.AddTable(tableName, table)
	ctx := sql.NewEmptyContext()

	rows := []sql.Row{
		sql.NewRow("a", "A", 1.0, 1),
		sql.NewRow("b", "B", 2.0, 2),
		sql.NewRow("c", "C", 3.0, 3),
		sql.NewRow("d", "D", 4.0, 4),
	}

	for _, row := range rows {
		table.Insert(ctx, row)
	}

	return db
}

func teardown(s *server.Server) {
	s.Close()
}

type Port string

// NewPort creates a new Port UDF.
func NewPort(port string) func(...sql.Expression) (sql.Expression, error) {
	return func(...sql.Expression) (sql.Expression, error) {
		return Port(port), nil
	}
}

func (f Port) FunctionName() string {
	return "rpc_port"
}

func (f Port) Type() sql.Type { return sql.LongText }

func (f Port) IsNullable() bool {
	return false
}

func (f Port) String() string {
	return "rpc_port()"
}

func (f Port) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(f, len(children), 0)
	}
	return f, nil
}

func (f Port) Resolved() bool {
	return true
}

func (f Port) Children() []sql.Expression { return nil }

func (f Port) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	if f == "" {
		return "2881", nil
	}
	return fmt.Sprintf("%s", string(f)), nil
}

type HostIp string

// NewHostIp creates a new HostIp UDF.
func NewHostIp(hostIp string) func(...sql.Expression) (sql.Expression, error) {
	return func(...sql.Expression) (sql.Expression, error) {
		return HostIp(hostIp), nil
	}
}

func (f HostIp) FunctionName() string {
	return "host_ip"
}

func (f HostIp) Type() sql.Type { return sql.LongText }

func (f HostIp) IsNullable() bool {
	return false
}

func (f HostIp) String() string {
	return "host_ip()"
}

func (f HostIp) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(f, len(children), 0)
	}
	return f, nil
}

func (f HostIp) Resolved() bool {
	return true
}

func (f HostIp) Children() []sql.Expression { return nil }

func (f HostIp) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	if f == "" {
		return "127.0.0.1", nil
	}
	return fmt.Sprintf("%s", string(f)), nil
}
