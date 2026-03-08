#!/bin/bash
export PORT=8080
./bin/gofi &
PID=$!
sleep 1
bombardier -c 125 -d 5s http://localhost:8080/middlewares &
curl -sK -v http://localhost:8080/debug/pprof/profile?seconds=5 > cpu.prof
kill $PID
go tool pprof -text cpu.prof | head -n 30
