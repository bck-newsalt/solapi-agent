package logger

import (
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Stdlog, Errlog *log.Logger

func Create(basePath string) error {
	// 콘솔 로그
	Stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// 파일 이름 및 코드 라인 출력 시 다음 추가:  log.LstdFlags | log.Lshortfile
	Errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
	Errlog.SetOutput(&lumberjack.Logger{
		Filename:   basePath + "/logs/agent.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	})

	Stdlog.Println("Logger Created!")

	return nil
}
