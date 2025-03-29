#!/bin/bash
#
set -e
[ -e ./bin ] || mkdir bin

go build -o ./bin/dirbackup ./cmd/dirbackup/main.go

go build -o ./bin/dirbackup-service ./cmd/dirbackup-service//main.go

go build -o ./bin/dirsynctime ./cmd/dirsynctime/main.go

docker buildx build -t dirbackup .
