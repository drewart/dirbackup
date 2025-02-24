#!/bin/bash
#
set -e
[ -e ./bin ] || mkdir bin

go build -o ./bin/dirbackup ./cmd/dirbackup/main.go

docker buildx build -t dirbackup .
