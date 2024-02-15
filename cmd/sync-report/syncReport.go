package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/bck-newsalt/solapi-agent/cmd/database"
	"github.com/bck-newsalt/solapi-agent/cmd/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/solapi/solapi-go"
)

type APIConfig struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"APISecret"`
	Protocol  string `json:"Protocol"`
	Domain    string `json:"Domain"`
	Prefix    string `json:"Prefix"`
	AppId     string `json:"AppId"`
}

var apiconf APIConfig

var client *solapi.Client

var basePath = "/opt/agent"

func readAPIConfig(homedir string, apiconf *APIConfig) error {
	var b []byte
	b, err := os.ReadFile(homedir + "/config.json")
	if err != nil {
		fmt.Println(err)
		return err
	}
	_ = json.Unmarshal(b, &apiconf)
	return nil
}

func syncMsgStatus(messageIds []string) {
	b, _ := json.Marshal(messageIds)
	params := make(map[string]string)
	params["messageIds[in]"] = string(b)
	params["limit"] = strconv.Itoa(len(messageIds))

	fmt.Println("메시지 상태 동기화:", len(messageIds), "건")

	result, err := client.Messages.GetMessageList(params)
	if err != nil {
		fmt.Println(err)
	}

	for _, res := range result.MessageList {
		_, err = database.DbImpl.Exec("UPDATE msg SET result = json_set(result, '$.status', ?, '$.statusCode', ?, '$.statusMessage', ?), updatedAt = NOW() WHERE messageId = ?", res.Status, res.StatusCode, res.Reason, res.MessageId)
		if err != nil {
			panic(err)
		}
	}
}

/*func printObj(obj interface{}) {
	var msgBytes []byte
	msgBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}
	msgStr := *(*string)(unsafe.Pointer(&msgBytes))
	fmt.Println(msgStr)
}*/

func main() {
	agentHome := os.Getenv("AGENT_HOME")
	if len(agentHome) > 0 {
		basePath = agentHome
	}

	var err error

	err = database.Connect(basePath)
	if err != nil {
		logger.Stdlog.Fatal(err)
	}

	err = readAPIConfig(basePath, &apiconf)
	if err != nil {
		panic(err)
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

	rows, err := database.DbImpl.Query("SELECT id, messageId FROM msg WHERE sent = true AND status != 'COMPLETE'")
	if err != nil {
		fmt.Println("DB Query ERROR:", err)
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rows)

	var id uint32
	var messageId string
	var messageIds []string
	for rows.Next() {
		_ = rows.Scan(&id, &messageId)
		messageIds = append(messageIds, messageId)
	}
	if len(messageIds) > 0 {
		syncMsgStatus(messageIds)
	}

	err = database.Close()
	if err != nil {
		logger.Stdlog.Fatal(err)
	}
}
