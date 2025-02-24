#!/bin/bash
#
set -e
[ -e ./bin ] || mkdir bin

export GOOS=linux 
export GOARCH=amd64
go build -o ./bin/dirbackup ./cmd/dirbackup/main.go

docker buildx build --platform linux/amd64 -t dirbackup .

