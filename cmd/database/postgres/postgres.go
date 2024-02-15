package postgres

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/bck-newsalt/solapi-agent/cmd/logger"
	"github.com/bck-newsalt/solapi-agent/cmd/types"
)

type PostgresSpec struct {
	db *sql.DB
}

func (s *PostgresSpec) Connect(dbconf types.DBConfig) error {

	var err error
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", dbconf.User, dbconf.Password, dbconf.Host, dbconf.Port, dbconf.DBName)
	logger.Errlog.Println("DB에 연결합니다 Connection String:", connectionString)
	// os.Getenv("DATABASE_URL")
	s.db, err = sql.Open("pgx", connectionString)
	if err != nil {
		logger.Errlog.Panicf("Unable to connect to database: %v\n", err)
	}
	// defer s.db.Close()

	var name string
	var weight int64
	// err = s.db.QueryRow("select name, weight from widgets where id=$1", 42).Scan(&name, &weight)
	err = s.db.QueryRow("select 1 from dual").Scan(&weight)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(name, weight)
	return nil
}

func (s *PostgresSpec) Exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *PostgresSpec) Query(query string, args ...any) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}
