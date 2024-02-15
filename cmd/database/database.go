package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"

	"github.com/bck-newsalt/solapi-agent/cmd/database/mysql"
	"github.com/bck-newsalt/solapi-agent/cmd/logger"
	"github.com/bck-newsalt/solapi-agent/cmd/types"
)

var Dbconf types.DBConfig
var DbImpl DBProviderImpl

func ReadDBConfig(basePath string) (string, error) {
	logger.Errlog.Printf("try, read %s/db.json\n", basePath)
	var b []byte
	b, err := os.ReadFile(basePath + "/db.json")
	if err != nil {
		logger.Errlog.Println(err)
		return "db.json 로딩 오류", err
	}

	err = json.Unmarshal(b, &Dbconf)
	if err != nil {
		logger.Errlog.Println(err)
		return "db.json parsing error", err
	}

	if Dbconf.Provider == "mysql" {
		return "", nil
	} else if Dbconf.Provider == "postgres" {
		return "", nil
	}

	return "", errors.New("bad DB config")
}

func Connect(basePath string) error {
	_, err := ReadDBConfig(basePath)
	if err != nil {
		logger.Stdlog.Fatal(err)
	}

	switch Dbconf.Provider {
	case "mysql":
		DbImpl = mysql.New()
		err = DbImpl.Connect(Dbconf)
		if err != nil {
			logger.Stdlog.Fatal(err)
		}
	case "postgres":
		break
	}
	return nil
}

type DBProviderImpl interface {
	Connect(types.DBConfig) error
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
}
