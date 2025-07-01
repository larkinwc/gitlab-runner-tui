.PHONY: build clean install run test lint

BINARY_NAME=gitlab-runner-tui
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/gitlab-runner-tui/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html coverage.txt
	@go clean

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete"

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-report:
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'

test-ci:
	@echo "Running tests for CI..."
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out

lint:
	@echo "Running linter..."
	@golangci-lint run || go vet ./...

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

dev:
	@echo "Running in development mode..."
	@go run cmd/gitlab-runner-tui/main.go