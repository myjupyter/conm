.PHONY: build smoke vet test lint tools fmt

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BIN ?= bin/$(GOOS)-$(GOARCH)/conm

TOOLS_BIN := $(CURDIR)/bin/tools
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint

INSTALL_TOOLS = GOBIN=$(TOOLS_BIN) go -C tools install tool

build:
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-s -w" -o $(BIN) .

smoke:
	@$(BIN) --help
	@$(BIN) --version

vet:
	@go vet ./...

test: vet
	@go test -race ./...

lint: $(GOLANGCI_LINT) vet
	@$(GOLANGCI_LINT) run ./...

tools:
	@$(INSTALL_TOOLS)

$(TOOLS_BIN)/%: tools/go.mod
	@$(INSTALL_TOOLS)

fmt: $(GOLANGCI_LINT)
	@go fix ./...
	@$(GOLANGCI_LINT) fmt ./...
