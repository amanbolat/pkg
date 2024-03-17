export GOBIN := $(PWD)/.bin
export PATH := $(GOBIN):$(PATH)
GO_PKGS=$(foreach pkg, $(shell go list ./...), $(if $(findstring /vendor/, $(pkg)), , $(pkg)))

all: mod lint test.all

.PHONY: mod
mod:
	go mod tidy
	cd thirdparty && go mod tidy
	go mod vendor

.PHONY: bin
bin:
	mkdir -p .bin

.PHONY: bin.go-enum
bin.go-enum: bin
	go install github.com/abice/go-enum

.PHONY: bin.golangci-lint
bin.golangci-lint: bin
	go install -modfile thirdparty/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: bin.golangci-lint
	golangci-lint run --fix

.PHONY: test.all
test.all:
	go test -count=1 ./...

.PHONY: gen.enums
gen.enums: bin.go-enum
	go-enum -file pkg/sql/migration_format.go --marshal --sql --nocase
