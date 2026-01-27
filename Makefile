.PHONY: all build test lint test-coverage clean help deps

all: help

help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run all tests with coverage"
	@echo "  test-coverage - Generate detailed coverage report"
	@echo "  lint          - Run linters"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"

build:
	go build -o speed-test main.go

test:
	go test -v -cover ./...

test-coverage:
	go test -coverprofile=coverage.out -covermode=count ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Open coverage.html in a browser to view detailed coverage."

lint:
	golangci-lint run ./...

clean:
	rm -f speed-test coverage.out coverage.html
	go clean

deps:
	go mod download
	go mod tidy
