.PHONY: build test clean install release docker-build

BINARY_NAME=baton
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-X baton/pkg/version.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./

test:
	go test -v ./...

integration-test:
	go test -v -tags=integration ./test/integration

e2e-test:
	go test -v -tags=e2e ./test/e2e

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -rf dist/

install: build
	cp $(BINARY_NAME) $(shell go env GOPATH)/bin/

release: clean
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./

docker-build:
	docker build -t baton:$(VERSION) .

mod-tidy:
	go mod tidy

mod-download:
	go mod download

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

all: fmt vet test build

dev: mod-tidy fmt vet test build

# Development helpers
run-init:
	./$(BINARY_NAME) init

run-status:
	./$(BINARY_NAME) status

run-start:
	./$(BINARY_NAME) start --dry-run

# Sample data for testing
sample-tasks:
	./$(BINARY_NAME) init
	./$(BINARY_NAME) ingest plan.md