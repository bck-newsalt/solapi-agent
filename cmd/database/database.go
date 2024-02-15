package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"

	"github.com/bck-newsalt/solapi-agent/cmd/database/mysql"
	"github.com/bck-newsalt/solapi-agent/cmd/database/postgres"
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
	case "postgres":
		DbImpl = postgres.New()
	}
	if DbImpl != nil {
		err = DbImpl.Connect(Dbconf)
		if err != nil {
			logger.Stdlog.Fatal(err)
		}
		logger.Stdlog.Println("Database Connected!")
	}
	return nil
}

func Close() error {
	var err error
	if DbImpl != nil {
		err = DbImpl.Close()
		if err != nil {
			logger.Stdlog.Fatal(err)
		}
		logger.Stdlog.Println("Database Closed!")
	}
	return err
}

type DBProviderImpl interface {
	Connect(types.DBConfig) error
	Close() error
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)

	FindLast1DayScheduled() (*sql.Rows, error)
	IncreseSendAttempts(id uint32) (sql.Result, error)
	UpdateComplete(messageId string, groupId string, status string, statusCode string, statusMessage string, id uint32) (sql.Result, error)

	FindLastReport() (*sql.Rows, error)
	FindPollReport() (*sql.Rows, error)
	IncreseReportAttempts(id uint32) (sql.Result, error)
	UpdateResultByMessageId(status string, statusCode string, reason string, messageId string) (sql.Result, error)
	UpdateFailed(statusCode string, statusMessage string, messageId string) (sql.Result, error)
}
