package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	logger "github.com/bck-newsalt/solapi-agent/cmd/logger"
	types "github.com/bck-newsalt/solapi-agent/cmd/types"
)

var Dbconf types.DBConfig
var Db *sql.DB

func getConnectionString(basePath string) (string, error) {
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
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", Dbconf.User, Dbconf.Password, Dbconf.Host, Dbconf.Port, Dbconf.DBName)
		return connectionString, nil
	} else if Dbconf.Provider == "postgres" {
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", Dbconf.User, Dbconf.Password, Dbconf.Host, Dbconf.Port, Dbconf.DBName)
		return connectionString, nil
	}

	return "", errors.New("bad DB config")
}

func Connect(basePath string) error {
	connectionString, err := getConnectionString(basePath)
	if err != nil {
		logger.Stdlog.Fatal(err)
	}
	logger.Errlog.Println("DB에 연결합니다 Connection String:", connectionString)

	Db, err = sql.Open(Dbconf.Provider, connectionString)
	if err != nil {
		panic(err)
	}
	Db.SetConnMaxLifetime(time.Minute * 3)
	Db.SetMaxOpenConns(10)
	Db.SetMaxIdleConns(10)
	return nil
}
