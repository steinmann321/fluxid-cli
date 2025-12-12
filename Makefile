.PHONY: build clean run test tools

# Build fluxid CLI into bin/
build:
	mkdir -p bin
	go build -o bin/fluxid ./cmd/fluxid

# Install required developer tools for hooks
tools:
	./hooks/setup_tools.sh

# Run the CLI
run:
	go run ./cmd/fluxid

# Run Go tests (including e2e)
test:
	go test ./...
# Clean artifacts
clean:
	rm -rf bin/
