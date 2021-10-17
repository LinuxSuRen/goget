run:
	go run cmd/server/root.go

build-client:
	go build -o goget cmd/cli/root.go
