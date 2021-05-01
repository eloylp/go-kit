.DEFAULT_GOAL := test

lint:
	golangci-lint run -v
lint-fix:
	golangci-lint run -v --fix
all: lint test
test: test-unit test-racy
test-unit:
	go test -v ./...
test-racy:
	go test -race -v --tags="racy" ./...
test-bench:
	go test -v -bench=. ./...