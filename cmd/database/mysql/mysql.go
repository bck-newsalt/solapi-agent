package mysql

import (
	"database/sql"
	"fmt"
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

	// ping
	if err != nil || s.db.Ping() != nil {
		logger.Errlog.Panicln(err.Error())
	}

	return nil
}

func (s *MysqlSpec) Close() error {
	err := s.db.Close()
	if err != nil {
		logger.Errlog.Fatalf("close db error: %v\n", err)
	}
	return err
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

func (s *MysqlSpec) FindLast1DayScheduled() (*sql.Rows, error) {
	return s.Query("SELECT id, payload FROM msg WHERE sent = false AND scheduledAt <= NOW() AND scheduledAt >= SUBDATE(NOW(), INTERVAL 24 HOUR) AND sendAttempts < 3 LIMIT 10000")
}

func (s *MysqlSpec) IncreseSendAttempts(id uint32) (sql.Result, error) {
	return s.Exec("UPDATE msg SET sendAttempts = sendAttempts + 1 WHERE id = ?", id)
}

func (s *MysqlSpec) UpdateComplete(messageId string, groupId string, status string, statusCode string, statusMessage string, id uint32) (sql.Result, error) {
	return s.Exec("UPDATE msg SET result = json_object('messageId', ?, 'groupId', ?, 'status', ?, 'statusCode', ?, 'statusMessage', ?), sent = true WHERE id = ?",
		messageId, groupId, status, statusCode, statusMessage, id)
}

func (s *MysqlSpec) FindLastReport() (*sql.Rows, error) {
	return s.Query("SELECT id, messageId, statusCode FROM msg WHERE sent = true AND createdAt < SUBDATE(NOW(), INTERVAL 72 HOUR) AND status != 'COMPLETE' LIMIT 100")
}

func (s *MysqlSpec) FindPollReport() (*sql.Rows, error) {
	return s.Query("SELECT id, messageId, statusCode FROM msg WHERE sent = true AND createdAt > SUBDATE(NOW(), INTERVAL 72 HOUR) AND updatedAt < SUBDATE(NOW(), INTERVAL (10 * (reportAttempts + 1)) SECOND) AND reportAttempts < 10 AND status != 'COMPLETE' LIMIT 100")
}

func (s *MysqlSpec) IncreseReportAttempts(id uint32) (sql.Result, error) {
	return s.Exec("UPDATE msg SET reportAttempts = reportAttempts + 1, updatedAt = NOW() WHERE id = ?", id)
}

func (s *MysqlSpec) UpdateResultByMessageId(status string, statusCode string, reason string, messageId string) (sql.Result, error) {
	return s.Exec("UPDATE msg SET result = json_set(result, '$.status', ?, '$.statusCode', ?, '$.statusMessage', ?), updatedAt = NOW() WHERE messageId = ?", status, statusCode, reason, messageId)
}

func (s *MysqlSpec) UpdateFailed(statusCode string, statusMessage string, messageId string) (sql.Result, error) {
	return s.Exec("UPDATE msg SET result = json_set(result, '$.status', 'COMPLETE', '$.statusCode', ?, '$.statusMessage', ?), updatedAt = NOW() WHERE messageId = ?", statusCode, statusMessage, messageId)
}
