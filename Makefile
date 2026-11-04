.PHONY: run test lint build clean tidy

# Run the API server
run:
	go run cmd/api/main.go

# Run all tests
test:
	go test ./... -v -count=1

# Run linter (requires golangci-lint installed)
lint:
	golangci-lint run ./...

# Build the binary
build:
	go build -o bin/api cmd/api/main.go

# Run the built binary
start: build
	./bin/api

# Clean build artifacts
clean:
	rm -rf bin/

# Tidy and verify dependencies
tidy:
	go mod tidy
	go mod verify

# Format all Go files
fmt:
	gofmt -s -w .

# Run all checks (format, lint, test)
check: fmt lint test
