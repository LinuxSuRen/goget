run:
	go run cmd/server/root.go
run-as-proxy:
	go run cmd/server/root.go --mode proxy --gc-duration 1s

build-server:
	CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o bin/goget-server cmd/server/root.go
build-client:
	go build -o bin/goget cmd/cli/root.go

goreleaser:
	goreleaser build --snapshot --rm-dist

build-image: build-server
	docker build bin -f Dockerfile -t ghcr.io/linuxsuren/goget-server:test
