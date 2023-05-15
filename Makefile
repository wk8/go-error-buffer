.DEFAULT_GOAL := all

.PHONY: all
all: test lint

# the TEST_FLAGS env var can be set to eg run only specific tests
.PHONY: test
test:
	go test -race -v -count=1 -cover $(TEST_FLAGS)

.PHONY: test_with_coverage
test_with_coverage: test
	go tool cover -html=cover.out

.PHONY: lint
lint:
	golangci-lint run
