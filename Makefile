.PHONY: build smoke vet test lint tools hooks fmt

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BIN ?= bin/$(GOOS)-$(GOARCH)/conm

TOOLS_BIN := $(CURDIR)/bin/tools
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint

INSTALL_TOOLS = GOBIN=$(TOOLS_BIN) go -C tools install tool

say = @printf '\033[1;36m▶\033[0m %s\n'

build:
	$(say) "build $(GOOS)/$(GOARCH) -> $(BIN)"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-s -w" -o $(BIN) .

smoke:
	$(say) "smoke $(BIN)"
	@$(BIN) --help
	@$(BIN) --version

vet:
	$(say) "vet"
	@go vet ./...

test: vet
	$(say) "test"
	@go test -race ./...

lint: $(GOLANGCI_LINT) vet
	$(say) "lint"
	@$(GOLANGCI_LINT) run ./...

tools:
	$(say) "tools -> $(TOOLS_BIN)"
	@$(INSTALL_TOOLS)

hooks: $(TOOLS_BIN)/lefthook
	$(say) "hooks install"
	@$(TOOLS_BIN)/lefthook install

$(TOOLS_BIN)/%: tools/go.mod
	$(say) "tools -> $(TOOLS_BIN)"
	@$(INSTALL_TOOLS)

fmt: $(GOLANGCI_LINT)
	$(say) "fmt"
	@go fix ./...
	@$(GOLANGCI_LINT) fmt ./...
