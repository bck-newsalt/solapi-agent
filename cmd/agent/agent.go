package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/bck-newsalt/solapi-agent/cmd/database"
	"github.com/bck-newsalt/solapi-agent/cmd/logger"
	"github.com/solapi/solapi-go"
	"github.com/takama/daemon"
)

type APIConfig struct {
	APIKey          string `json:"apiKey"`
	APISecret       string `json:"APISecret"`
	Protocol        string `json:"Protocol"`
	Domain          string `json:"Domain"`
	Prefix          string `json:"Prefix"`
	AppId           string `json:"AppId"`
	AllowDuplicates bool   `json:"AllowDuplicates"`
}

const (
	name        = "solapi"
	description = "Solapi Agent Service"
)

var apiconf APIConfig

var client *solapi.Client

var basePath string = "/opt/agent"

// Service has embedded daemon
type Service struct {
	daemon.Daemon
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	usage := "Usage: solapi install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	logger.Stdlog.Println("데몬 시작")
	logger.Errlog.Println("데몬 시작")

	// loop work cycle with accept connections or interrupt
	// by system signal

	var err error

	err = readAPIConfig(basePath, &apiconf)
	if err != nil {
		logger.Stdlog.Fatal(err)
	}

	client = solapi.NewClient()
	client.Messages.Config = map[string]string{
		"APIKey":    apiconf.APIKey,
		"APISecret": apiconf.APISecret,
		"Protocol":  apiconf.Protocol,
		"Domain":    apiconf.Domain,
		"Prefix":    apiconf.Prefix,
	}
	client.Storage.Config = map[string]string{
		"APIKey":    apiconf.APIKey,
		"APISecret": apiconf.APISecret,
		"Protocol":  apiconf.Protocol,
		"Domain":    apiconf.Domain,
		"Prefix":    apiconf.Prefix,
	}

	err = database.Connect(basePath)
	if err != nil {
		logger.Stdlog.Fatal(err)
	}

	// create DB
	if database.Dbconf.Provider == "postgres" {
		_, err = database.DbImpl.Exec(`
			CREATE SCHEMA IF NOT EXISTS sms AUTHORIZATION postgres;
			CREATE TABLE IF NOT EXISTS sms.msg (
			id SERIAL4 NOT NULL,
			createdAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updatedAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			scheduledAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			sendAttempts SMALLINT NOT NULL DEFAULT 0,
			reportAttempts SMALLINT NOT NULL DEFAULT 0,
			"to" VARCHAR(20) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(payload, 'to'))) STORED,
			"from" VARCHAR(20) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(payload, 'from'))) STORED,
			groupId VARCHAR(255) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'groupId'))) STORED,
			messageId VARCHAR(255) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'messageId'))) STORED,
			status VARCHAR(20) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'status'))) STORED,
			statusCode VARCHAR(255) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'statusCode'))) STORED,
			statusMessage VARCHAR(255) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'statusMessage'))) STORED,
			payload JSON,
			result JSON DEFAULT NULL,
			sent BOOLEAN NOT NULL DEFAULT false,
			CONSTRAINT msg_pkey PRIMARY KEY (id)
			);
			CREATE INDEX ix_msg_id ON sms.msg USING btree (id);
			CREATE INDEX ix_msg_createdAt ON sms.msg USING btree (createdAt);
			CREATE INDEX ix_msg_updatedAt ON sms.msg USING btree (updatedAt);
			CREATE INDEX ix_msg_scheduledAt ON sms.msg USING btree (scheduledAt);
			CREATE INDEX ix_msg_sendAttempts ON sms.msg USING btree (sendAttempts);
			CREATE INDEX ix_msg_reportAttempts ON sms.msg USING btree (reportAttempts);
			CREATE INDEX ix_msg_to ON sms.msg USING btree ("to");
			CREATE INDEX ix_msg_from ON sms.msg USING btree ("from");
			CREATE INDEX ix_msg_groupId ON sms.msg USING btree (groupId);
			CREATE INDEX ix_msg_messageId ON sms.msg USING btree (messageId);
			CREATE INDEX ix_msg_status ON sms.msg USING btree (status);
			CREATE INDEX ix_msg_statusCode ON sms.msg USING btree (statusCode);
			CREATE INDEX ix_msg_sent ON sms.msg USING btree (sent);	
			`)
		if err != nil {
			logger.Stdlog.Fatal(err)
		}
	}

	// go 키워드로 함수를 호출하면, goroutine 실행

	go pollMsg()
	go pollResult()
	go pollLastReport()

	// monitor mem
	// go func() {
	// 	var rtm runtime.MemStats
	// 	for {
	// 		runtime.GC()
	// 		time.Sleep(time.Millisecond * 300)
	// 		runtime.ReadMemStats(&rtm)
	// 		logger.Errlog.Println("Allow:", rtm.Alloc)
	// 		logger.Errlog.Println("TotalAllow:", rtm.TotalAlloc)
	// 		logger.Errlog.Println("Sys:", rtm.Sys)
	// 		logger.Errlog.Println("Mallocs:", rtm.Mallocs)
	// 		logger.Errlog.Println("Frees:", rtm.Frees)
	// 	}
	// }()

	// for-select
	// for {
	// 	select {
	// receive from channel
	// 	case killSignal := <-interrupt:
	// 		logger.Errlog.Println("시스템 시그널이 감지되었습니다:", killSignal)
	// 		if killSignal == os.Interrupt {
	// 			return "Daemon was interrupted by system signal", nil
	// 		}
	// 		return "Daemon was killed", nil
	// 	}
	// }

	// for-range, if channel closed,
	for killSignal := range interrupt {
		logger.Errlog.Println("시스템 시그널이 감지되었습니다:", killSignal)
		err = database.Close()
		if err != nil {
			logger.Stdlog.Fatal(err)
		}
		if killSignal == os.Interrupt {
			return "Daemon was interrupted by system signal", nil
		}
		return "Daemon was killed", nil
	}

	return usage, nil
}

func readAPIConfig(homedir string, apiconf *APIConfig) error {
	var b []byte
	b, err := os.ReadFile(homedir + "/config.json")
	if err != nil {
		logger.Errlog.Println(err)
		return err
	}
	_ = json.Unmarshal(b, &apiconf)
	return nil
}

func pollMsg() {
	logger.Stdlog.Println("pollMsg - loop begin")
	// loop
	for {
		// wait 1sec
		time.Sleep(time.Second * 1)
		rows, err := database.DbImpl.FindLast1DayScheduled()
		if err != nil {
			logger.Errlog.Println("[메시지발송] DB Query ERROR:", err)
			time.Sleep(time.Second * 5)
			continue
		}

		// logger.Stdlog.Println("FindLast1DayScheduled rows:", rows)

		var messageList []interface{}
		var groupId string

		var id uint32
		var count = 0
		var idList []uint32
		for rows.Next() {
			// logger.Stdlog.Println("pollMsg begin process 1message")
			var payload string
			var msgObj map[string]interface{}
			if count == 0 {
				params := make(map[string]string)
				if apiconf.AllowDuplicates == true {
					params["allowDuplicates"] = "true"
				}
				result, err := client.Messages.CreateGroup(params)
				if err != nil {
					logger.Errlog.Printf("CreateGroup error: %v\n", err)
				}
				logger.Stdlog.Printf("CreateGroup result: %v\n", result)
				groupId = result.GroupId
			}
			count++

			err := rows.Scan(&id, &payload)
			if err != nil {
				logger.Errlog.Printf("row Scan error: %v\n", err)
			}
			logger.Stdlog.Printf("pollMsg - has message to send id: %d payload: %s\n", id, payload)
			// _ = fmt.Sprintf("id: %u", id)
			_, err = database.DbImpl.IncreseSendAttempts(id)
			if err != nil {
				logger.Errlog.Printf("update sendAttempts error: %v\n", err)
				continue
			}
			err = json.Unmarshal([]byte(payload), &msgObj)
			if err != nil {
				logger.Errlog.Printf("Unmarshal error: %v\n", err)
				continue
			}

			file := msgObj["file"]
			if file != nil {
				// upload file
				filename := file.(string)
				var fullpath string
				if strings.HasPrefix(filename, "/") {
					fullpath = filename
				} else {
					fullpath = basePath + "/files/" + filename
				}
				logger.Errlog.Println("파일 업로드:", fullpath)
				params := make(map[string]string)
				params["file"] = fullpath
				params["name"] = "customFileName"
				params["type"] = "MMS"
				result, err := client.Storage.UploadFile(params)
				if err != nil {
					logger.Errlog.Println(err)
					continue
				}
				logger.Errlog.Println("이미지 아이디:", result.FileId)
				msgObj["imageId"] = result.FileId
				delete(msgObj, "file")
			}

			messageList = append(messageList, msgObj)
			idList = append(idList, id)
		}

		if err := rows.Err(); err != nil {
			logger.Errlog.Fatalln("[메시지발송] DB rows.Err() ERROR:", err)
		}

		err = rows.Close()
		if err != nil {
			logger.Errlog.Fatalln("[메시지발송] DB rows.Close() ERROR:", err)
		}

		if len(messageList) > 0 {
			var msgParams = make(map[string]interface{})
			msgParams["messages"] = messageList

			result, err := client.Messages.AddGroupMessage(groupId, msgParams)
			if err != nil {
				logger.Errlog.Printf("AddGroupMessage error: %v\n", err)
				continue
			}
			for i, res := range result.ResultList {
				status := "COMPLETE"
				if res.StatusCode == "2000" {
					status = "PENDING"
				}
				_, err = database.DbImpl.UpdateComplete(res.MessageId, groupId, status, res.StatusCode, res.StatusMessage, idList[i])
				if err != nil {
					logger.Errlog.Println("UpdateComplete error:", err)
					continue
				}
			}

			groupInfo, err2 := client.Messages.SendGroup(groupId)
			if err2 != nil {
				logger.Errlog.Printf("SendGroup error: %v\n", err2)
				continue
			}
			printObj(groupInfo.Count)
		}
	}
}

func pollLastReport() {
	for {
		time.Sleep(time.Second * 1)
		rows, err := database.DbImpl.FindLastReport()
		if err != nil {
			logger.Errlog.Println("[마지막 리포트] DB Query ERROR:", err)
			time.Sleep(time.Second * 60)
			continue
		}

		var id uint32
		var messageId string
		var statusCode string
		var messageIds []string
		for rows.Next() {
			_ = rows.Scan(&id, &messageId, &statusCode)
			messageIds = append(messageIds, messageId)
		}
		if len(messageIds) > 0 {
			syncMsgStatus(messageIds, statusCode, "3040")
		}

		_ = rows.Close()
	}
}

func pollResult() {
	for {
		time.Sleep(time.Millisecond * 500)
		rows, err := database.DbImpl.FindPollReport()
		if err != nil {
			logger.Errlog.Println("[리포트 처리] DB Query ERROR:", err)
			time.Sleep(time.Second * 10)
			continue
		}

		var id uint32
		var messageId string
		var statusCode string
		var messageIds []string
		for rows.Next() {
			_ = rows.Scan(&id, &messageId, &statusCode)

			_, err = database.DbImpl.IncreseReportAttempts(id)
			messageIds = append(messageIds, messageId)
		}
		if len(messageIds) > 0 {
			syncMsgStatus(messageIds, statusCode, "")
		}

		_ = rows.Close()
	}
}

func syncMsgStatus(messageIds []string, statusCode string, defaultCode string) {
	b, _ := json.Marshal(messageIds)
	params := make(map[string]string)
	params["messageIds[in]"] = string(b)
	params["limit"] = strconv.Itoa(len(messageIds))

	logger.Errlog.Println("메시지 상태 동기화:", len(messageIds), "건")

	result, err := client.Messages.GetMessageList(params)
	if err != nil {
		logger.Errlog.Println(err)
	}

	for _, res := range result.MessageList {
		if res.StatusCode != statusCode {
			_, err = database.DbImpl.UpdateResultByMessageId(res.Status, res.StatusCode, res.Reason, res.MessageId)
			if err != nil {
				panic(err)
			}
		} else {
			if defaultCode != "" {
				_, err = database.DbImpl.UpdateFailed(defaultCode, "전송시간 초과", res.MessageId)
			}
		}
	}
}

func printObj(obj interface{}) {
	// logger.Errlog.Println("printObj:")
	var msgBytes []byte
	msgBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}
	msgStr := *(*string)(unsafe.Pointer(&msgBytes))
	logger.Errlog.Println(msgStr)
}

func init() {
	// base path
	agentHome := os.Getenv("AGENT_HOME")
	if len(agentHome) > 0 {
		basePath = agentHome
	} else {
		// pwd
		dir, err := os.Getwd()
		if err != nil {
			log.Panicf("Getwd error: %v\n", err)
		}
		// fmt.Printf("pwd: %s\n", dir)
		basePath = dir
	}
	log.Printf("init basePath: %s\n", basePath)

	// logger
	logger.Create(basePath)
}

func main() {

	daemonType := daemon.SystemDaemon
	goos := runtime.GOOS
	switch goos {
	case "windows":
		daemonType = daemon.SystemDaemon
		logger.Stdlog.Println("OS: Windows")
	case "darwin":
		daemonType = daemon.UserAgent
		logger.Stdlog.Println("OS: MAC")
	case "linux":
		daemonType = daemon.SystemDaemon
		logger.Stdlog.Println("OS: Linux")
	default:
		fmt.Printf("%s.\n", goos)
	}

	srv, err := daemon.New(name, description, daemonType)
	if err != nil {
		logger.Errlog.Println("Error: ", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		logger.Errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	logger.Stdlog.Println(status)
}
