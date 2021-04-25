.DEFAULT_GOAL := test

lint:
	golangci-lint run -v
lint-fix:
	golangci-lint run -v --fix
all: lint test
test: test-unit test-integration test-racy
test-unit:
	go test -race -v --tags="unit" ./...
test-integration:
	go test -race -v --tags="integration" ./...
test-racy:
	go test -race -v --tags="racy" ./...
test-bench:
	go test -v -bench=. ./...