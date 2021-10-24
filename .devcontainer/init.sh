#!/bin/bash

curl -L https://github.com/linuxsuren/http-downloader/releases/latest/download/hd-linux-amd64.tar.gz | tar xzv
sudo mv hd /usr/bin/hd
go run cmd/server/root.go --externalAddress $(npx codespaces-port 7878 -q)

