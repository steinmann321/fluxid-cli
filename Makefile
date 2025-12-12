.PHONY: build clean test run install

# Build the fluxid binary
build:
	mkdir -p bin
	go build -o bin/fluxid .

# Install binary to PATH (optional)
install: build
	cp bin/fluxid $(shell go env GOPATH)/bin/

# Run the application
run:
	go run .

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Default target
all: build
