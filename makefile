# Makefile
.PHONY: help build test clean install dev release-local

BINARY_NAME := docker-machine-driver-hetzner
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X main.version=$(VERSION) \
           -X main.gitCommit=$(COMMIT) \
           -X main.buildDate=$(DATE)

BUILD_FLAGS := -ldflags "$(LDFLAGS)"
RELEASE_FLAGS := -ldflags "-s -w $(LDFLAGS)" -trimpath

help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)

build-all:
	@echo "Building for all platforms..."
	GOOS=linux   GOARCH=amd64 go build $(RELEASE_FLAGS) -o bin/$(BINARY_NAME)-linux-amd64
	GOOS=linux   GOARCH=arm64 go build $(RELEASE_FLAGS) -o bin/$(BINARY_NAME)-linux-arm64
	
test:
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/ coverage.out

install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/
	@echo "Testing installation..."
	$(BINARY_NAME) -v

dev: clean build install

release-local:
	goreleaser release --snapshot --clean

release-check:
	goreleaser check

docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

deps:
	go mod tidy
	go mod verify

fmt:
	gofmt -s -w .
	goimports -w .

.DEFAULT_GOAL := help
