This project aims to provide a way to get binary file from a Golang project easily. Users don't need to have a Golang 
environment.

## Server

Usage:
```shell
docker run --restart always -d -v /var/data/goget:/tmp -p 7878:7878 ghcr.io/linuxsuren/goget-server:latest
```

## Client

Usage:
```shell
goget github.com/linuxsuren/http-downloader
```
