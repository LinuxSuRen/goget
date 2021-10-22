This project aims to provide a way to get binary file from a Golang project easily. Users don't need to have a Golang 
environment.

## Server

Usage:
```shell
docker run --restart always -d -v /var/data/goget:/tmp -p 7878:7878 ghcr.io/linuxsuren/goget-server:latest
```

## Client

Simple usage:
```shell
goget github.com/linuxsuren/http-downloader
```

Non standard go project usage:
```shell
goget gitee.com/linuxsuren/goget --package cmd/cli/root.go
```

## Other HTTP clients

You can use any kinds of HTTP clients to get your desired binary file. Such as, use curl command to download it:

```shell
curl http://localhost:7878/gitee.com/linuxsuren/http-downloader --output hd && chmod u+x hd
```

Get more details about the [API](doc/api.md).
