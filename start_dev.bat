@echo off
start cmd /k "cd ml-service && uvicorn main:app --reload --port 5002"
echo Starting Go Server...
go run main.go
