PROJECT_NAME := $(shell basename "$(PWD)")
GO_LINT_CI_VERSION := v1.27.0
GO_LINT_CI_PATH := $(shell go env GOPATH)/bin

.DEFAULT_GOAL := test

lint:
	golangci-lint run -v
lint-fix:
	golangci-lint run -v --fix
linter-install:
	wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_LINT_CI_PATH) $(GO_LINT_CI_VERSION)
all: lint test

test: test-unit test-integration test-bench test-racy

test-unit:
	go test -race -v --tags="unit" ./...
test-integration:
	go test -race -v --tags="integration" ./...
test-racy:
	go test -race -v --tags="racy" ./...
test-bench:
	go test -v -bench=. ./...