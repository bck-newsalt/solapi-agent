@echo off

SETLOCAL

SET GOOS=linux
SET GOARCH=amd64
go build -o build/agent-linux-amd64 ./cmd/agent/agent.go

ENDLOCAL
