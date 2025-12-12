.PHONY: build clean run test tools

# Build hello-world binary and fluxid CLI into bin/
build:
	mkdir -p bin
	go build -o bin/hello ./src
	go build -o bin/fluxid ./cmd/fluxid

# Install required developer tools for hooks
tools:
	./hooks/setup_tools.sh

# Run the app
run:
	go run ./src

# Run Go tests (including e2e)
test:
	go test ./...

# Clean artifacts
clean:
	rm -rf bin/
