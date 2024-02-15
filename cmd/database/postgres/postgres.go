package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bck-newsalt/solapi-agent/cmd/logger"
	"github.com/bck-newsalt/solapi-agent/cmd/types"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresSpec struct {
	db *sql.DB
}

var instance PostgresSpec

func New() *PostgresSpec {
	instance = PostgresSpec{}
	return &instance
}

func Dispose() error {
	return instance.db.Close()
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

	s.db.SetConnMaxLifetime(time.Minute * 3)
	s.db.SetMaxOpenConns(10)
	s.db.SetMaxIdleConns(10)

	return nil
}

func (s *PostgresSpec) Close() error {
	err := s.db.Close()
	if err != nil {
		logger.Errlog.Fatalf("close db error: %v\n", err)
	}
	return err
}

func (s *PostgresSpec) Exec(query string, args ...any) (sql.Result, error) {
	// logger.Stdlog.Printf("Exec() - query: %s args: %d", query, len(args))
	if len(args) > 0 {
		return s.db.Exec(query, args...)
	}
	return s.db.Exec(query)
}

func (s *PostgresSpec) Query(query string, args ...any) (*sql.Rows, error) {
	// logger.Stdlog.Printf("Exec() - query: %s args: %d", query, len(args))
	if len(args) > 0 {
		return s.db.Query(query, args...)
	}
	return s.db.Query(query)
}

func (s *PostgresSpec) FindLast1DayScheduled() (*sql.Rows, error) {
	return s.Query("SELECT id, payload FROM sms.msg WHERE sent = false AND scheduledAt <= NOW() AND scheduledAt >= (NOW() - INTERVAL '24 HOURS') AND sendAttempts < 3 LIMIT 10000")
}

func (s *PostgresSpec) IncreseSendAttempts(id uint32) (sql.Result, error) {
	return s.Exec("UPDATE sms.msg SET sendAttempts = sendAttempts + 1 WHERE id = ?", id)
}

func (s *PostgresSpec) UpdateComplete(messageId string, groupId string, status string, statusCode string, statusMessage string, id uint32) (sql.Result, error) {
	return s.Exec("UPDATE sms.msg SET result = jsonb_build_object('messageId', ?, 'groupId', ?, 'status', ?, 'statusCode', ?, 'statusMessage', ?), sent = true WHERE id = ?",
		messageId, groupId, status, statusCode, statusMessage, id)
}

func (s *PostgresSpec) FindLastReport() (*sql.Rows, error) {
	return s.Query("SELECT id, messageId, statusCode FROM sms.msg WHERE sent = true AND createdAt < (NOW() - INTERVAL '72 HOURS') AND status != 'COMPLETE' LIMIT 100")
}

func (s *PostgresSpec) FindPollReport() (*sql.Rows, error) {
	return s.Query("SELECT id, messageId, statusCode FROM sms.msg WHERE sent = true AND createdAt > (NOW() - INTERVAL '72 HOURS') AND updatedAt < (NOW() - MAKE_INTERVAL(SECS => 10 * (reportAttempts + 1))) AND reportAttempts < 10 AND status != 'COMPLETE' LIMIT 100")
}

func (s *PostgresSpec) IncreseReportAttempts(id uint32) (sql.Result, error) {
	return s.Exec("UPDATE sms.msg SET reportAttempts = reportAttempts + 1, updatedAt = NOW() WHERE id = ?", id)
}

func (s *PostgresSpec) UpdateResultByMessageId(status string, statusCode string, reason string, messageId string) (sql.Result, error) {
	return s.Exec("UPDATE sms.msg SET result = coalesce(result, '{}'::json)::jsonb || jsonb_build_object('status', ?) || jsonb_build_object('statusCode', ?) || jsonb_build_object('statusMessage', ?), updatedAt = NOW() WHERE messageId = ?", status, statusCode, reason, messageId)
}

func (s *PostgresSpec) UpdateFailed(statusCode string, statusMessage string, messageId string) (sql.Result, error) {
	return s.Exec("UPDATE sms.msg SET result = coalesce(result, '{}'::json)::jsonb || jsonb_build_object('status', 'COMPLETE') || jsonb_build_object('statusCode', ?) || jsonb_build_object('statusMessage', ?), updatedAt = NOW() WHERE messageId = ?", statusCode, statusMessage, messageId)
}
