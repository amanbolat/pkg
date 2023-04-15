export GOBIN := $(PWD)/.bin
export PATH := $(GOBIN):$(PATH)
export GOFLAGS := -mod=mod
GO_PKGS=$(foreach pkg, $(shell go list ./...), $(if $(findstring /vendor/, $(pkg)), , $(pkg)))

all: mod lint test.all

.PHONY: mod
mod:
	go mod tidy
	cd thirdparty && go mod tidy

.PHONY: bin
bin:
	mkdir -p .bin

.PHONY: bin.golangci-lint
bin.golangci-lint: bin
	go install -modfile thirdparty/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: bin.golangci-lint
	golangci-lint run --fix

.PHONY: test.all
test.all:
	go test -count=1 ./...
