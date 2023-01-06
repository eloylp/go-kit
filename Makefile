.DEFAULT_GOAL := test

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: lint-fix
lint-fix:
	golangci-lint run -v --fix

.PHONY: all
all: lint test

.PHONY: test
test: test-unit test-racy

.PHONY: test-unit
test-unit:
	go test -tags unit -v ./...

.PHONY: test-racy
test-racy:
	go test -tags racy -count=1 -v -race ./...

.PHONY: bench
bench:
	go test -v -bench=. ./...