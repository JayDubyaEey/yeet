.PHONY: build test lint clean install run help

# Variables
BINARY_NAME=yeet
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/JayDubyaEey/yeet/pkg/version.Version=$(VERSION)"

# Default target
help:
	@echo "Available targets:"
	@echo "  build      Build the binary"
	@echo "  test       Run tests"
	@echo "  lint       Run linters"
	@echo "  clean      Clean build artifacts"
	@echo "  install    Install binary to /usr/local/bin"
	@echo "  run        Run the binary"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	@which golangci-lint > /dev/null || echo "golangci-lint not installed"
	golangci-lint run ./... || true

fmt:
	gofmt -s -w .
	go mod tidy

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/

run: build
	./$(BINARY_NAME)
