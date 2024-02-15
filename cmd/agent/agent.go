package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/bck-newsalt/solapi-agent/cmd/database"
	logger "github.com/bck-newsalt/solapi-agent/cmd/logger"
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

	go pollMsg()
	go pollResult()
	go pollLastReport()
	/*
	  go func() {
	    var rtm runtime.MemStats
	    for {
	      runtime.GC()
	      time.Sleep(time.Millisecond * 300)
	      runtime.ReadMemStats(&rtm)
	      logger.Errlog.Println("Allow:", rtm.Alloc)
	      logger.Errlog.Println("TotalAllow:", rtm.TotalAlloc)
	      logger.Errlog.Println("Sys:", rtm.Sys)
	      logger.Errlog.Println("Mallocs:", rtm.Mallocs)
	      logger.Errlog.Println("Frees:", rtm.Frees)
	    }
	  }()
	*/

	for {
		select {
		case killSignal := <-interrupt:
			logger.Errlog.Println("시스템 시그널이 감지되었습니다:", killSignal)
			if killSignal == os.Interrupt {
				return "Daemon was interrupted by system signal", nil
			}
			return "Daemon was killed", nil
		}
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
	for {
		time.Sleep(time.Second * 1)
		rows, err := database.Db.Query("SELECT id, payload FROM msg WHERE sent = false AND scheduledAt <= NOW() AND scheduledAt >= SUBDATE(NOW(), INTERVAL 24 HOUR) AND sendAttempts < 3 LIMIT 10000")
		if err != nil {
			logger.Errlog.Println("[메시지발송] DB Query ERROR:", err)
			time.Sleep(time.Second * 5)
			continue
		}

		var messageList []interface{}
		var groupId string

		var id uint32
		var count = 0
		var idList []uint32
		for rows.Next() {
			var payload string
			var msgObj map[string]interface{}
			if count == 0 {
				params := make(map[string]string)
				if apiconf.AllowDuplicates == true {
					params["allowDuplicates"] = "true"
				}
				result, err := client.Messages.CreateGroup(params)
				if err != nil {
					logger.Errlog.Println(err)
				}
				groupId = result.GroupId
			}
			count++

			err := rows.Scan(&id, &payload)
			if err != nil {
				logger.Errlog.Println(err)
			}
			_ = fmt.Sprintf("id: %u", id)
			_, err = database.Db.Exec("UPDATE msg SET sendAttempts = sendAttempts + 1 WHERE id = ?", id)
			if err != nil {
				logger.Errlog.Println(err)
				continue
			}
			err = json.Unmarshal([]byte(payload), &msgObj)
			if err != nil {
				logger.Errlog.Println(err)
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
		_ = rows.Close()
		if len(messageList) > 0 {
			var msgParams = make(map[string]interface{})
			msgParams["messages"] = messageList

			result, err := client.Messages.AddGroupMessage(groupId, msgParams)
			if err != nil {
				logger.Errlog.Println(err)
				continue
			}
			for i, res := range result.ResultList {
				status := "COMPLETE"
				if res.StatusCode == "2000" {
					status = "PENDING"
				}
				_, err = database.Db.Exec("UPDATE msg SET result = json_object('messageId', ?, 'groupId', ?, 'status', ?, 'statusCode', ?, 'statusMessage', ?), sent = true WHERE id = ?", res.MessageId, groupId, status, res.StatusCode, res.StatusMessage, idList[i])
				if err != nil {
					logger.Errlog.Println(err)
					continue
				}
			}

			groupInfo, err2 := client.Messages.SendGroup(groupId)
			if err2 != nil {
				logger.Errlog.Println(err2)
				continue
			}
			printObj(groupInfo.Count)
		}
	}
}

func pollLastReport() {
	for {
		time.Sleep(time.Second * 1)
		rows, err := database.Db.Query("SELECT id, messageId, statusCode FROM msg WHERE sent = true AND createdAt < SUBDATE(NOW(), INTERVAL 72 HOUR) AND status != 'COMPLETE' LIMIT 100")
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
		rows, err := database.Db.Query("SELECT id, messageId, statusCode FROM msg WHERE sent = true AND createdAt > SUBDATE(NOW(), INTERVAL 72 HOUR) AND updatedAt < SUBDATE(NOW(), INTERVAL (10 * (reportAttempts + 1)) SECOND) AND reportAttempts < 10 AND status != 'COMPLETE' LIMIT 100")
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

			_, err = database.Db.Exec("UPDATE msg SET reportAttempts = reportAttempts + 1, updatedAt = NOW() WHERE id = ?", id)
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
			_, err = database.Db.Exec("UPDATE msg SET result = json_set(result, '$.status', ?, '$.statusCode', ?, '$.statusMessage', ?), updatedAt = NOW() WHERE messageId = ?", res.Status, res.StatusCode, res.Reason, res.MessageId)
			if err != nil {
				panic(err)
			}
		} else {
			if defaultCode != "" {
				_, err = database.Db.Exec("UPDATE msg SET result = json_set(result, '$.status', 'COMPLETE', '$.statusCode', ?, '$.statusMessage', ?), updatedAt = NOW() WHERE messageId = ?", defaultCode, "전송시간 초과", res.MessageId)
			}
		}
	}
}

func printObj(obj interface{}) {
	var msgBytes []byte
	msgBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}
	msgStr := *(*string)(unsafe.Pointer(&msgBytes))
	logger.Errlog.Println(msgStr)
}

func init() {
	agentHome := os.Getenv("AGENT_HOME")
	if len(agentHome) > 0 {
		basePath = agentHome
	}

	logger.Create(basePath)
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		logger.Stdlog.Fatal(err)
	}
	fmt.Printf("pwd: %s\n", dir)

	daemonType := daemon.SystemDaemon
	goos := runtime.GOOS
	switch goos {
	case "windows":
		daemonType = daemon.SystemDaemon
		logger.Stdlog.Println("OS: Windows")
		basePath = dir
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
