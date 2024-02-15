package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/bck-newsalt/solapi-agent/cmd/logger"
	"github.com/bck-newsalt/solapi-agent/cmd/types"
	_ "github.com/go-sql-driver/mysql"
)

// database.DBProviderImpl 구현 필요
type MysqlSpec struct {
	db *sql.DB
}

var instance MysqlSpec

func New() *MysqlSpec {
	instance = MysqlSpec{}
	return &instance
}

func (s *MysqlSpec) Connect(dbconf types.DBConfig) error {
	var err error
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbconf.User, dbconf.Password, dbconf.Host, dbconf.Port, dbconf.DBName)
	logger.Errlog.Println("DB에 연결합니다 Connection String:", connectionString)
	s.db, err = sql.Open("mysql", connectionString)
	if err != nil {
		logger.Errlog.Panicf("Unable to connect to database: %v\n", err)
	}
	// defer s.db.Close()

	s.db.SetConnMaxLifetime(time.Minute * 3)
	s.db.SetMaxOpenConns(10)
	s.db.SetMaxIdleConns(10)

	var name string
	var weight int64
	// err = s.db.QueryRow("select name, weight from widgets where id=$1", 42).Scan(&name, &weight)
	err = s.db.QueryRow("select 9876 from dual").Scan(&weight)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("name: %s, weight: %d\n", name, weight)

	return nil
}

func (s *MysqlSpec) Exec(query string, args ...any) (sql.Result, error) {
	// logger.Stdlog.Printf("Exec() - query: %s args: %d", query, len(args))
	if len(args) > 0 {
		return s.db.Exec(query, args...)
	}
	return s.db.Exec(query)
}

func (s *MysqlSpec) Query(query string, args ...any) (*sql.Rows, error) {
	// logger.Stdlog.Printf("Query() - query: %s args: %d", query, len(args))
	if len(args) > 0 {
		return s.db.Query(query, args...)
	}
	return s.db.Query(query)
}
