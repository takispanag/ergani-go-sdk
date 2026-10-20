GO ?= go
BINARY ?= ergani-app
COVERAGE ?= coverage.out

.PHONY: all build test test-race cover vet fmt fmt-check lint readme tidy check clean

all: check

## build: compile the package and the example binary
build:
	$(GO) build ./...
	$(GO) build -o $(BINARY) .

## test: run the test suite
test:
	$(GO) test ./...

## test-race: run the test suite with the race detector
test-race:
	$(GO) test -race ./...

## cover: run the test suite and report the coverage
cover:
	$(GO) test -covermode=atomic -coverprofile=$(COVERAGE) ./...
	$(GO) tool cover -func=$(COVERAGE)

## vet: run go vet
vet:
	$(GO) vet ./...

## fmt: format the sources
fmt:
	gofmt -w .

## fmt-check: fail when a source file is not gofmt'd
fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt'd:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

## lint: run golangci-lint, when it is installed
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint is not installed, skipping"; exit 0; }
	golangci-lint run

## readme: compile every Go example of the README
readme:
	$(GO) run scripts/check_readme.go

## tidy: tidy up go.mod
tidy:
	$(GO) mod tidy

## check: run every check the CI runs
check: fmt-check vet test-race lint readme

## clean: remove the build artefacts
clean:
	rm -f $(BINARY) $(COVERAGE)
