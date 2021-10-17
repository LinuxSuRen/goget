run:
	go run cmd/server/root.go

build-client:
	go build -o bin/goget cmd/cli/root.go

goreleaser:
	goreleaser build --snapshot --rm-dist
