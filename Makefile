.PHONY: build clean run test tools

# Build hello-world binary into bin/
build:
	mkdir -p bin
	go build -o bin/hello ./src

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
