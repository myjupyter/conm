.PHONY: build lint fmt

build:
	@go build -o bin/conm main.go

lint:
	@golangci-lint run ./...

fmt:
	@go fix ./...
	@golangci-lint fmt ./...
